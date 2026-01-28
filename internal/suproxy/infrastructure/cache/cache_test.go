package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedRepoMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func newTestCacheService(mockRepo *sharedRepoMocks.MockCache) *cacheService {
	return &cacheService{
		log:  newTestLogger(),
		cfg:  &config.Suproxy{},
		repo: mockRepo,
		ttls: entity.CacheTTLPolicy{
			SupplierOK:   10 * time.Minute,
			MockOK:       20 * time.Minute,
			ErrorOrEmpty: 45 * time.Second,
		},
		tracer: otel.Tracer("test"),
	}
}

func TestCacheService_Lookup_Miss(t *testing.T) {
	mockRepo := sharedRepoMocks.NewMockCache(t)
	svc := newTestCacheService(mockRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{"X-Test": "1"},
		Body:        `{"foo":"bar"}`,
	}

	key := svc.BuildKey(req, false)

	mockRepo.
		On("Get", mock.Anything, key).
		Return([]byte(nil), false, nil)

	resp, hit, err := svc.Lookup(context.Background(), req, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hit {
		t.Fatalf("expected cache miss, got hit")
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty response on miss, got %q", string(resp))
	}

	mockRepo.AssertExpectations(t)
}

func TestCacheService_Lookup_Hit(t *testing.T) {
	mockRepo := sharedRepoMocks.NewMockCache(t)
	svc := newTestCacheService(mockRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{"X-Test": "1"},
		Body:        `{"foo":"bar"}`,
	}

	key := svc.BuildKey(req, false)
	cachedResp := json.RawMessage(`{"ok":true}`)
	entry := entity.CacheEntry{
		Mock:     false,
		Key:      key,
		Request:  req,
		Response: cachedResp,
		CachedAt: time.Now().Add(-1 * time.Minute).UTC(),
		Version:  1,
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("failed to marshal cache entry: %v", err)
	}

	mockRepo.
		On("Get", mock.Anything, key).
		Return(payload, true, nil)

	resp, hit, err := svc.Lookup(context.Background(), req, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !hit {
		t.Fatalf("expected cache hit, got miss")
	}
	if string(resp) != string(cachedResp) {
		t.Fatalf("expected response %q, got %q", string(cachedResp), string(resp))
	}

	mockRepo.AssertExpectations(t)
}

func TestCacheService_Lookup_RepoErrorDegradesToMiss(t *testing.T) {
	mockRepo := sharedRepoMocks.NewMockCache(t)
	svc := newTestCacheService(mockRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{},
		Body:        "",
	}

	key := svc.BuildKey(req, false)

	mockRepo.
		On("Get", mock.Anything, key).
		Return([]byte(nil), false, errors.New("redis down"))

	resp, hit, err := svc.Lookup(context.Background(), req, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hit {
		t.Fatalf("expected miss on repo error")
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty response on repo error, got %q", string(resp))
	}

	mockRepo.AssertExpectations(t)
}

func TestCacheService_Store_UsesCorrectTTL(t *testing.T) {
	cases := []struct {
		name        string
		isMock      bool
		expectedTTL time.Duration
		reqDest     string
	}{
		{
			name:        "supplier_OK",
			isMock:      false,
			expectedTTL: 10 * time.Minute,
			reqDest:     "/supplier",
		},
		{
			name:        "mock_OK",
			isMock:      true,
			expectedTTL: 20 * time.Minute,
			reqDest:     "/mock",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := sharedRepoMocks.NewMockCache(t)
			svc := newTestCacheService(mockRepo)

			req := entity.Request{
				Destination: tc.reqDest,
				Header:      map[string]string{},
				Body:        "{}",
			}
			response := []byte(`{"result":"ok"}`)

			var usedTTL time.Duration

			mockRepo.
				On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					ttl, ok := args.Get(3).(time.Duration)
					if !ok {
						t.Fatalf("expected ttl time.Duration, got %T", args.Get(3))
					}
					usedTTL = ttl
				}).
				Return(nil)

			err := svc.Store(context.Background(), req, response, tc.isMock, false)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if usedTTL != tc.expectedTTL {
				t.Fatalf("expected TTL %v, got %v", tc.expectedTTL, usedTTL)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCacheService_Store_UsesCorrectTTL_ForErrorOrEmpty(t *testing.T) {
	cases := []struct {
		name    string
		isError bool
		resp    []byte
	}{
		{"error_flag", true, []byte(`{"result":"ok"}`)},
		{"empty_object", false, []byte("{}")},
		{"empty_array", false, []byte("[]")},
		{"null", false, []byte("null")},
		{"empty_string", false, []byte("")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := sharedRepoMocks.NewMockCache(t)
			svc := newTestCacheService(mockRepo)

			req := entity.Request{
				Destination: "/test",
				Header:      map[string]string{},
				Body:        "{}",
			}

			var usedTTL time.Duration

			mockRepo.
				On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					ttl, ok := args.Get(3).(time.Duration)
					if !ok {
						t.Fatalf("expected ttl to be time.Duration, got %T", args.Get(3))
					}
					usedTTL = ttl
				}).
				Return(nil)

			err := svc.Store(context.Background(), req, tc.resp, false, tc.isError)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if usedTTL != svc.ttls.ErrorOrEmpty {
				t.Fatalf("expected TTL %v, got %v", svc.ttls.ErrorOrEmpty, usedTTL)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCacheService_Invalidate_UsesCorrectKey(t *testing.T) {
	mockRepo := sharedRepoMocks.NewMockCache(t)
	svc := newTestCacheService(mockRepo)

	req := entity.Request{
		Destination: "/invalidate",
		Header:      map[string]string{"X": "1"},
		Body:        "body",
	}

	expectedKey := svc.BuildKey(req, false)

	mockRepo.
		On("Delete", mock.Anything, expectedKey).
		Return(nil)

	err := svc.Invalidate(context.Background(), req, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	mockRepo.AssertExpectations(t)
}

func TestBuildKey_IsDeterministicAndHeaderOrderIndependent(t *testing.T) {
	mockRepo := sharedRepoMocks.NewMockCache(t)
	svc := newTestCacheService(mockRepo)

	req1 := entity.Request{
		Destination: "/search",
		Header: map[string]string{
			"X-B": "2",
			"X-A": "1",
		},
		Body: "body",
	}

	req2 := entity.Request{
		Destination: "/search",
		Header: map[string]string{
			"X-A": "1",
			"X-B": "2",
		},
		Body: "body",
	}

	key1 := svc.BuildKey(req1, false)
	key2 := svc.BuildKey(req2, false)

	if key1 != key2 {
		t.Fatalf("expected keys to be equal, got %q and %q", key1, key2)
	}
}

func TestIsLikelyEmptyJSON(t *testing.T) {
	cases := []struct {
		input    []byte
		expected bool
	}{
		{[]byte(""), true},
		{[]byte("   "), true},
		{[]byte("{}"), true},
		{[]byte("[]"), true},
		{[]byte("null"), true},
		{[]byte(`{"foo":"bar"}`), false},
		{[]byte(`[1,2,3]`), false},
	}

	for _, tc := range cases {
		got := isLikelyEmptyJSON(tc.input)
		if got != tc.expected {
			t.Fatalf("isLikelyEmptyJSON(%q) = %v, expected %v", string(tc.input), got, tc.expected)
		}
	}
}
