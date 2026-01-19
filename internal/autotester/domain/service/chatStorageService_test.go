package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	servmocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	cfg := &config.Config{
		DefaultPageSize: 10,
	}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	emptySummaries := []*entity.ChatSummary{}
	someSummaries := []*entity.ChatSummary{
		{ChatId: "chat1", UpdatedAt: time.Unix(100, 0)},
		{ChatId: "chat2", UpdatedAt: time.Unix(200, 0)},
	}

	tests := []struct {
		name        string
		logger      *slog.Logger
		listReturns []any
		wantErr     bool
	}{
		{
			name:        "all not nil - empty summaries",
			logger:      logger,
			listReturns: []any{emptySummaries, nil},
			wantErr:     false,
		},
		{
			name:        "all not nil - with summaries",
			logger:      logger,
			listReturns: []any{someSummaries, nil},
			wantErr:     false,
		},
		{
			name:    "nil assertion error",
			logger:  nil,
			wantErr: true,
		},
		{
			name:        "repo.ListAll returns error",
			logger:      logger,
			listReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockValidator := servmocks.NewMockValidator(t)

			if test.listReturns != nil {
				mockRepo.On("ListAll", mock.Anything).Return(test.listReturns...)
			}

			svc, err := NewChatStorageService(test.logger, mockRepo, mockValidator, tracer, cfg)
			if (err != nil) != test.wantErr {
				t.Errorf("NewChatStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewChatStorageService() returned nil service")
			}
			if test.listReturns != nil {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// nolint: dupl
func TestChatStorageSaveChat(t *testing.T) {
	cfg := &config.Config{
		DefaultPageSize: 10,
	}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	existingSummaries := []*entity.ChatSummary{
		{ChatId: "existing1", UpdatedAt: time.Unix(100, 0), Groups: []string{"group1"}},
	}

	tests := []struct {
		name             string
		initialSummary   []*entity.ChatSummary
		createReturns    []any
		validRetuns      []any
		wantErr          bool
		chat             *entity.Chat
		expectNewSummary bool
	}{
		{
			name:             "success - new chat",
			initialSummary:   []*entity.ChatSummary{},
			createReturns:    []any{nil},
			validRetuns:      []any{nil},
			wantErr:          false,
			chat:             &entity.Chat{Id: "chat1", Title: "Test", Groups: []string{"g1"}, UpdatedAt: time.Unix(300, 0)},
			expectNewSummary: true,
		},
		{
			name:             "success - update existing chat",
			initialSummary:   existingSummaries,
			createReturns:    []any{nil},
			validRetuns:      []any{nil},
			wantErr:          false,
			chat:             &entity.Chat{Id: "existing1", Title: "Updated", Groups: []string{"g2"}, UpdatedAt: time.Unix(400, 0)},
			expectNewSummary: false,
		},
		{
			name:    "nil assert error",
			wantErr: true,
			chat:    nil,
		},
		{
			name:           "validation error",
			initialSummary: []*entity.ChatSummary{},
			wantErr:        true,
			validRetuns:    []any{errors.New("err")},
			chat:           &entity.Chat{},
		},
		{
			name:           "repo returns error",
			initialSummary: []*entity.ChatSummary{},
			createReturns:  []any{errors.New("repo error")},
			validRetuns:    []any{nil},
			wantErr:        true,
			chat:           &entity.Chat{Id: "chat2", Groups: []string{"g1"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return(test.initialSummary, nil).Maybe()

			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.chat, mock.AnythingOfType("*entity.ChatSummary")).Return(test.createReturns...)
			}
			if test.validRetuns != nil {
				mockVal.On("ValidateChat", mock.Anything, test.chat).Return(test.validRetuns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, tracer, cfg)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			err = svc.SaveChat(context.Background(), test.chat)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveChat() error = %v, wantErr %v", err, test.wantErr)
			}
			mockRepo.AssertExpectations(t)
			mockVal.AssertExpectations(t)
		})
	}
}

func TestChatStorageLoadChat(t *testing.T) {
	cfg := &config.Config{
		DefaultPageSize: 10,
	}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name         string
		chatid       string
		loadReturns  []any
		validReturns []any
		ctx          context.Context
		wantErr      bool
	}{
		{
			name:         "success",
			chatid:       "chat123",
			loadReturns:  []any{&entity.Chat{}, nil},
			validReturns: []any{nil},
			ctx:          context.Background(),
			wantErr:      false,
		},
		{
			name:    "nil assert failed",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "invalid chatId",
			chatid:  "",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:        "repo returns error",
			chatid:      "chat123",
			loadReturns: []any{nil, errors.New("repo error")},
			ctx:         context.Background(),
			wantErr:     true,
		},
		{
			name:         "repo returns invalid chat",
			chatid:       "chat123",
			loadReturns:  []any{&entity.Chat{}, nil},
			validReturns: []any{errors.New("err")},
			ctx:          context.Background(),
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return([]*entity.ChatSummary{}, nil).Maybe()

			if test.loadReturns != nil {
				mockRepo.On("Read", mock.Anything, test.chatid).Return(test.loadReturns...)
			}
			if test.validReturns != nil {
				mockVal.On("ValidateChat", mock.Anything, &entity.Chat{}).Return(test.validReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, tracer, cfg)
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
			mockRepo.AssertExpectations(t)
			mockVal.AssertExpectations(t)
		})
	}
}

//nolint:funlen
func TestLoadSummaries(t *testing.T) {
	cfg := &config.Config{
		DefaultPageSize: 10,
	}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	allSummaries := []*entity.ChatSummary{
		{ChatId: "chat5", UpdatedAt: time.Unix(500, 0), Groups: []string{"1", "2"}},
		{ChatId: "chat4", UpdatedAt: time.Unix(400, 0), Groups: []string{"1"}},
		{ChatId: "chat3", UpdatedAt: time.Unix(300, 0), Groups: []string{"2"}},
		{ChatId: "chat2", UpdatedAt: time.Unix(200, 0), Groups: []string{"1"}},
		{ChatId: "chat1", UpdatedAt: time.Unix(100, 0), Groups: []string{"3"}},
	}

	tests := []struct {
		name            string
		groups          []string
		offset          int
		limit           int
		initialSummary  []*entity.ChatSummary
		expected        []*entity.ChatSummary
		expectedHasMore bool
		wantErr         bool
		ctx             context.Context
	}{
		{
			name:            "success - no groups, no pagination",
			offset:          0,
			limit:           0,
			initialSummary:  allSummaries,
			expected:        allSummaries,
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
		{
			name:            "success - with limit",
			offset:          0,
			limit:           2,
			initialSummary:  allSummaries,
			expected:        allSummaries[0:2],
			expectedHasMore: true,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
		{
			name:            "success - with offset and limit",
			offset:          2,
			limit:           2,
			initialSummary:  allSummaries,
			expected:        allSummaries[2:4],
			expectedHasMore: true,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
		{
			name:            "success - last page",
			offset:          3,
			limit:           10,
			initialSummary:  allSummaries,
			expected:        allSummaries[3:5],
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
		{
			name:            "success - filtered by group",
			offset:          0,
			limit:           10,
			initialSummary:  allSummaries,
			expected:        []*entity.ChatSummary{allSummaries[0], allSummaries[1], allSummaries[3]},
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{"1"},
		},
		{
			name:            "success - filtered by group with pagination",
			offset:          0,
			limit:           2,
			initialSummary:  allSummaries,
			expected:        []*entity.ChatSummary{allSummaries[0], allSummaries[1]},
			expectedHasMore: true,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{"1"},
		},
		{
			name:            "success - filtered by multiple groups",
			offset:          0,
			limit:           10,
			initialSummary:  allSummaries,
			expected:        []*entity.ChatSummary{allSummaries[0], allSummaries[1], allSummaries[2], allSummaries[3]},
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{"1", "2"},
		},
		{
			name:           "nil context error",
			wantErr:        true,
			ctx:            nil,
			initialSummary: allSummaries,
		},
		{
			name:           "offset negative",
			wantErr:        true,
			ctx:            context.Background(),
			offset:         -1,
			initialSummary: allSummaries,
		},
		{
			name:           "empty group string error",
			wantErr:        true,
			ctx:            context.Background(),
			groups:         []string{""},
			initialSummary: allSummaries,
		},
		{
			name:           "offset too large - no groups",
			offset:         10,
			limit:          5,
			initialSummary: allSummaries,
			wantErr:        true,
			ctx:            context.Background(),
			groups:         []string{},
		},
		{
			name:           "offset too large - with groups",
			offset:         10,
			limit:          5,
			initialSummary: allSummaries,
			wantErr:        true,
			ctx:            context.Background(),
			groups:         []string{"1"},
		},
		{
			name:            "empty summaries list",
			offset:          0,
			limit:           10,
			initialSummary:  []*entity.ChatSummary{},
			expected:        []*entity.ChatSummary{},
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
		{
			name:            "offset at boundary",
			offset:          4,
			limit:           10,
			initialSummary:  allSummaries,
			expected:        allSummaries[4:5],
			expectedHasMore: false,
			wantErr:         false,
			ctx:             context.Background(),
			groups:          []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return(test.initialSummary, nil)

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, tracer, cfg)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}

			res, hasMore, err := svc.LoadSummaries(test.ctx, test.offset, test.limit, test.groups...)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
				assert.False(t, hasMore)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, res)
				assert.Equal(t, test.expectedHasMore, hasMore)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
