package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	servmocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	mockRepo := mocks.NewMockChatStorageRepository(t)
	mockValidator := servmocks.NewMockValidator(t)
	mockCache := servmocks.NewMockCache(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil assertion error",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewChatStorageService(test.logger, mockRepo, mockValidator, mockCache, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewSessionSummaryStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewSessionSummaryStorageService() returned nil service")
			}
		})
	}
}

// nolint: dupl
func TestChatStorageSaveChat(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name           string
		createReturns  []any
		validRetuns    []any
		mockCacheSetup []any
		wantErr        bool
		chat           *entity.Chat
	}{
		{
			name:           "success - storing in s3 and cache",
			createReturns:  []any{nil},
			validRetuns:    []any{nil},
			mockCacheSetup: []any{nil},
			wantErr:        false,
			chat:           &entity.Chat{},
		},
		{
			name:           "success - storing in s3 but cache fails",
			createReturns:  []any{nil},
			validRetuns:    []any{nil},
			mockCacheSetup: []any{fmt.Errorf("cache storing error")},
			wantErr:        false,
			chat:           &entity.Chat{},
		},
		{
			name:    "nil assert error",
			wantErr: true,
			chat:    nil,
		},
		{
			name:        "error - validation error",
			wantErr:     true,
			validRetuns: []any{errors.New("err")},
			chat:        &entity.Chat{},
		},
		{
			name:          "error - repo returns error",
			createReturns: []any{errors.New("repo error")},
			validRetuns:   []any{nil},
			wantErr:       true,
			chat:          &entity.Chat{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)
			mockCache := servmocks.NewMockCache(t)

			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.chat).Return(test.createReturns...)
			}
			if test.validRetuns != nil {
				mockVal.On("ValidateChat", mock.Anything, test.chat).Return(test.validRetuns...)
			}

			if test.mockCacheSetup != nil {
				mockCache.On("Store", mock.Anything, test.chat).Return(test.mockCacheSetup...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			err = svc.SaveChat(context.Background(), test.chat)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveChat() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// nolint:funlen
func TestChatStorageLoadChat(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name                 string
		chatid               string
		loadReturns          []any
		validReturns         []any
		mockSetupLookupCache []any
		mockSetupStoreCache  []any
		ctx                  context.Context
		wantErr              bool
	}{
		{
			name:                 "success - loading from s3, cache miss, cache store success",
			chatid:               "chat123",
			loadReturns:          []any{&entity.Chat{}, nil},
			validReturns:         []any{nil},
			mockSetupLookupCache: []any{nil, nil},
			mockSetupStoreCache:  []any{nil},
			ctx:                  context.Background(),
			wantErr:              false,
		},
		{
			name:                 "success - loading from s3, cache miss, cache store error",
			chatid:               "chat123",
			loadReturns:          []any{&entity.Chat{}, nil},
			validReturns:         []any{nil},
			mockSetupLookupCache: []any{nil, nil},
			mockSetupStoreCache:  []any{fmt.Errorf("cache store error")},
			ctx:                  context.Background(),
			wantErr:              false,
		},
		{
			name:    "nil assert failed",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "empty chatId",
			chatid:  "",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:                 "repo returns error, cache miss",
			chatid:               "chat123",
			loadReturns:          []any{nil, errors.New("repo error")},
			mockSetupLookupCache: []any{nil, fmt.Errorf("cache error")},
			ctx:                  context.Background(),
			wantErr:              true,
		},
		{
			name:                 "repo returns invalid chat",
			chatid:               "chat123",
			loadReturns:          []any{&entity.Chat{}, nil},
			validReturns:         []any{errors.New("err")},
			mockSetupLookupCache: []any{nil, nil},
			ctx:                  context.Background(),
			wantErr:              true,
		},
		{
			name:                 "cache hit but invalid chat",
			chatid:               "chat123",
			mockSetupLookupCache: []any{&entity.Chat{}, nil},
			validReturns:         []any{errors.New("err")},
			loadReturns:          []any{&entity.Chat{}, nil},
			ctx:                  context.Background(),
			wantErr:              true,
		},
		{
			name:                 "cache hit with valid chat",
			chatid:               "chat123",
			mockSetupLookupCache: []any{&entity.Chat{}, nil},
			validReturns:         []any{nil},
			ctx:                  context.Background(),
			wantErr:              false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)
			mockCache := servmocks.NewMockCache(t)

			if test.loadReturns != nil {
				mockRepo.On("Read", mock.Anything, test.chatid).Return(test.loadReturns...)
			}
			if test.validReturns != nil {
				mockVal.On("ValidateChat", mock.Anything, &entity.Chat{}).Return(test.validReturns...)
			}
			if test.mockSetupLookupCache != nil {
				mockCache.On("LookUp", mock.Anything, mock.Anything).Return(test.mockSetupLookupCache...)
			}
			if test.mockSetupStoreCache != nil {
				mockCache.On("Store", mock.Anything, mock.Anything).Return(test.mockSetupStoreCache...)
			}

			// mockCache.On("LookUp", mock.Anything, test.ChatId).Return(tt.setupCacheMock...)
			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadChat(test.ctx, test.chatid)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestLoadSummaries(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	orderedResult := []*entity.ChatSummary{{UpdatedAt: time.Unix(200, 0), Groups: []string{"1", "2"}}, {UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}}}

	tests := []struct {
		name        string
		groups      []string
		listReturns []any
		wantErr     bool
		ctx         context.Context
	}{
		{
			name:        "success, no groups",
			listReturns: []any{orderedResult, nil},
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{},
		},
		{
			name:        "success, filtered by Group groups",
			listReturns: []any{append(orderedResult, &entity.ChatSummary{UpdatedAt: time.Unix(1, 0), Groups: []string{"0"}}), nil},
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{"1"},
		},
		{
			name:    "nil assert error",
			wantErr: true,
			ctx:     nil,
		},
		{
			name:    "group assert error",
			wantErr: true,
			ctx:     context.Background(),
			groups:  []string{""},
		},
		{
			name:        "inverse order",
			listReturns: []any{[]*entity.ChatSummary{{UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}}, {UpdatedAt: time.Unix(200, 0), Groups: []string{"1", "2"}}}, nil},
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{},
		},
		{
			name:        "repo returns error",
			listReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
			ctx:         context.Background(),
			groups:      []string{},
		},
	}

	mockCache := servmocks.NewMockCache(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.listReturns != nil {
				mockRepo.On("ListAll", mock.Anything).Return(test.listReturns...)
			}

			val := servmocks.NewMockValidator(t)

			svc, err := NewChatStorageService(logger, mockRepo, val, mockCache, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadSummaries(test.ctx, test.groups...)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, orderedResult, res)
			}
		})
	}
}
