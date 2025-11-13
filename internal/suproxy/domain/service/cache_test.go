package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// fakeCacheRepository is a simple test double for CacheRepository
type fakeCacheRepository struct {
	getFunc    func(ctx context.Context, key string) ([]byte, bool, error)
	setFunc    func(ctx context.Context, key string, value []byte, ttl time.Duration) error
	deleteFunc func(ctx context.Context, key string) error

	lastSetKey   string
	lastSetValue []byte
	lastSetTTL   time.Duration

	lastDeleteKey string
}

// Get simulates a cache read using the configured test callback
func (f *fakeCacheRepository) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if f.getFunc != nil {
		return f.getFunc(ctx, key)
	}
	return nil, false, nil
}

// Set simulates a cache write and records the last written key, value and TTL
func (f *fakeCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	f.lastSetKey = key
	f.lastSetValue = value
	f.lastSetTTL = ttl

	if f.setFunc != nil {
		return f.setFunc(ctx, key, value, ttl)
	}
	return nil
}

// Delete simulates deleting a cache entry and records the last deleted key
func (f *fakeCacheRepository) Delete(ctx context.Context, key string) error {
	f.lastDeleteKey = key
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, key)
	}
	return nil
}

// Close simulates closing the cache connection (no-op for fake)
func (f *fakeCacheRepository) Close() error {
	return nil
}

// newTestLogger creates a logger that discards all output for testing
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// newTestCacheService creates a cacheService with a fake repository and fixed TTLs
func newTestCacheService(repo repository.Cache) *cacheService {
	return &cacheService{
		log:  newTestLogger(),
		cfg:  nil,
		repo: repo,
		ttls: entity.CacheTTLPolicy{
			SupplierOK:   10 * time.Minute,
			MockOK:       20 * time.Minute,
			ErrorOrEmpty: 45 * time.Second,
		},
	}
}

// TestCacheService_Lookup_Miss verifies Lookup returns a cache miss when no entry is stored
func TestCacheService_Lookup_Miss(t *testing.T) {
	fakeRepo := &fakeCacheRepository{
		getFunc: func(ctx context.Context, key string) ([]byte, bool, error) {
			return nil, false, nil
		},
	}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{"X-Test": "1"},
		Request:     `{"foo":"bar"}`,
	}

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
}

// TestCacheService_Lookup_Hit verifies Lookup returns the cached response on a cache hit
func TestCacheService_Lookup_Hit(t *testing.T) {
	fakeRepo := &fakeCacheRepository{}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{"X-Test": "1"},
		Request:     `{"foo":"bar"}`,
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

	fakeRepo.getFunc = func(ctx context.Context, k string) ([]byte, bool, error) {
		if k != key {
			t.Fatalf("expected key %q, got %q", key, k)
		}
		return payload, true, nil
	}

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
}

// TestCacheService_Lookup_RepoErrorDegradesToMiss ensures repo errors are treated as cache misses
func TestCacheService_Lookup_RepoErrorDegradesToMiss(t *testing.T) {
	fakeRepo := &fakeCacheRepository{
		getFunc: func(ctx context.Context, key string) ([]byte, bool, error) {
			return nil, false, errors.New("redis down")
		},
	}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/test",
		Header:      map[string]string{},
		Request:     "",
	}

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
}

// TestCacheService_Store_UsesCorrectTTL_ForSupplierOK checks TTL for successful supplier responses
func TestCacheService_Store_UsesCorrectTTL_ForSupplierOK(t *testing.T) {
	fakeRepo := &fakeCacheRepository{}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/supplier",
		Header:      map[string]string{},
		Request:     "{}",
	}
	response := []byte(`{"result":"ok"}`)

	err := svc.Store(context.Background(), req, response, false, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fakeRepo.lastSetTTL != svc.ttls.SupplierOK {
		t.Fatalf("expected TTL %v, got %v", svc.ttls.SupplierOK, fakeRepo.lastSetTTL)
	}
}

// TestCacheService_Store_UsesCorrectTTL_ForMockOK checks TTL for successful mock responses
func TestCacheService_Store_UsesCorrectTTL_ForMockOK(t *testing.T) {
	fakeRepo := &fakeCacheRepository{}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/mock",
		Header:      map[string]string{},
		Request:     "{}",
	}
	response := []byte(`{"result":"ok"}`)

	err := svc.Store(context.Background(), req, response, true, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fakeRepo.lastSetTTL != svc.ttls.MockOK {
		t.Fatalf("expected TTL %v, got %v", svc.ttls.MockOK, fakeRepo.lastSetTTL)
	}
}

// TestCacheService_Store_UsesCorrectTTL_ForErrorOrEmpty verifies ErrorOrEmpty TTL for error/empty responses
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeRepo := &fakeCacheRepository{}
			svc := newTestCacheService(fakeRepo)

			req := entity.Request{
				Destination: "/test",
				Header:      map[string]string{},
				Request:     "{}",
			}

			err := svc.Store(context.Background(), req, tc.resp, false, tc.isError)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if fakeRepo.lastSetTTL != svc.ttls.ErrorOrEmpty {
				t.Fatalf("expected TTL %v, got %v", svc.ttls.ErrorOrEmpty, fakeRepo.lastSetTTL)
			}
		})
	}
}

// TestCacheService_Invalidate_UsesCorrectKey ensures Invalidate deletes the entry with the correct key
func TestCacheService_Invalidate_UsesCorrectKey(t *testing.T) {
	fakeRepo := &fakeCacheRepository{}
	svc := newTestCacheService(fakeRepo)

	req := entity.Request{
		Destination: "/invalidate",
		Header:      map[string]string{"X": "1"},
		Request:     "body",
	}

	expectedKey := svc.BuildKey(req, false)

	err := svc.Invalidate(context.Background(), req, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if fakeRepo.lastDeleteKey != expectedKey {
		t.Fatalf("expected delete key %q, got %q", expectedKey, fakeRepo.lastDeleteKey)
	}
}

// TestBuildKey_IsDeterministicAndHeaderOrderIndependent verifies BuildKey is deterministic and order-insensitive
func TestBuildKey_IsDeterministicAndHeaderOrderIndependent(t *testing.T) {
	svc := newTestCacheService(&fakeCacheRepository{})

	req1 := entity.Request{
		Destination: "/search",
		Header: map[string]string{
			"X-B": "2",
			"X-A": "1",
		},
		Request: "body",
	}

	req2 := entity.Request{
		Destination: "/search",
		Header: map[string]string{
			"X-A": "1",
			"X-B": "2",
		},
		Request: "body",
	}

	key1 := svc.BuildKey(req1, false)
	key2 := svc.BuildKey(req2, false)

	if key1 != key2 {
		t.Fatalf("expected keys to be equal, got %q and %q", key1, key2)
	}
}

// TestIsLikelyEmptyJSON validates the helper that detects empty/null-like JSON responses
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
