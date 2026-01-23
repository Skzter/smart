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
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
)

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	mockCache := servmocks.NewMockCache(t)
	tracer := otel.Tracer("test")
	mockMetrics := sharedMocks.NewMockMetricsService(t)
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
			svc, err := NewChatStorageService(test.logger, mockRepo, mockValidator, mockCache, tracer, mockMetrics)

			if (err != nil) != test.wantErr {
				t.Errorf("NewChatSummaryStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewChatSummaryStorageService() returned nil service")
			}
		})
	}
}

//nolint:funlen
func TestChatStorageSaveChat(t *testing.T) {
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
		mockCacheSetup   []any
		wantErr          bool
		chat             *entity.Chat
		expectNewSummary bool
	}{
		{
			name:             "success - storing in s3 and cache",
			initialSummary:   []*entity.ChatSummary{},
			createReturns:    []any{nil},
			validRetuns:      []any{nil},
			mockCacheSetup:   []any{nil},
			wantErr:          false,
			chat:             &entity.Chat{Id: "chat1", Title: "Test", Groups: []string{"g1"}, UpdatedAt: time.Unix(300, 0)},
			expectNewSummary: true,
		},
		{
			name:             "success - update existing chat",
			initialSummary:   existingSummaries,
			createReturns:    []any{nil},
			validRetuns:      []any{nil},
			mockCacheSetup:   []any{nil},
			wantErr:          false,
			chat:             &entity.Chat{Id: "existing1", Title: "Updated", Groups: []string{"g2"}, UpdatedAt: time.Unix(400, 0)},
			expectNewSummary: false,
		},
		{
			name:             "success - storing in s3 but cache fails",
			createReturns:    []any{nil},
			validRetuns:      []any{nil},
			mockCacheSetup:   []any{fmt.Errorf("cache storing error")},
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
			name:           "error - validation error",
			initialSummary: []*entity.ChatSummary{},
			wantErr:        true,
			validRetuns:    []any{errors.New("err")},
			chat:           &entity.Chat{},
		},
		{
			name:           "error - repo returns error",
			initialSummary: []*entity.ChatSummary{},
			createReturns:  []any{errors.New("repo error")},
			validRetuns:    []any{nil},
			wantErr:        true,
			chat:           &entity.Chat{Id: "chat2", Groups: []string{"g1"}},
		},
	}
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncCacheHit").Return().Maybe()
	mockMetricsServ.On("IncCacheMiss").Return().Maybe()
	mockMetricsServ.On("RecordCacheDuration", mock.Anything, mock.Anything).Return().Maybe()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)
			mockCache := servmocks.NewMockCache(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return(test.initialSummary, nil).Maybe()

			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.chat, mock.AnythingOfType("*entity.ChatSummary")).Return(test.createReturns...)
			}
			if test.validRetuns != nil {
				mockVal.On("ValidateChat", mock.Anything, test.chat).Return(test.validRetuns...)
			}

			if test.mockCacheSetup != nil {
				mockCache.On("Store", mock.Anything, test.chat).Return(test.mockCacheSetup...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer, mockMetricsServ)
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

// nolint:funlen
// TestChatStorageSaveChatSortingAndInsertion tests the binary search insertion logic
// and verifies that summaries remain sorted after insertions and updates
func TestChatStorageSaveChatSortingAndInsertion(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name             string
		initialSummaries []*entity.ChatSummary
		chatsToSave      []*entity.Chat
		expectedOrder    []string // Expected ChatIds in order after all saves
		mockCacheSetup   []any
	}{
		{
			name: "insert at beginning - most recent timestamp",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chat2", UpdatedAt: time.Unix(200, 0), Groups: []string{"g1"}},
				{ChatId: "chat1", UpdatedAt: time.Unix(100, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chat3", Title: "Newest", Groups: []string{"g1"}, UpdatedAt: time.Unix(300, 0)},
			},
			expectedOrder:  []string{"chat3", "chat2", "chat1"},
			mockCacheSetup: []any{nil},
		},
		{
			name: "insert at end - oldest timestamp",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chat3", UpdatedAt: time.Unix(300, 0), Groups: []string{"g1"}},
				{ChatId: "chat2", UpdatedAt: time.Unix(200, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chat1", Title: "Oldest", Groups: []string{"g1"}, UpdatedAt: time.Unix(100, 0)},
			},
			expectedOrder:  []string{"chat3", "chat2", "chat1"},
			mockCacheSetup: []any{nil},
		},
		{
			name: "insert in middle",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chat3", UpdatedAt: time.Unix(300, 0), Groups: []string{"g1"}},
				{ChatId: "chat1", UpdatedAt: time.Unix(100, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chat2", Title: "Middle", Groups: []string{"g1"}, UpdatedAt: time.Unix(200, 0)},
			},
			expectedOrder:  []string{"chat3", "chat2", "chat1"},
			mockCacheSetup: []any{nil},
		},
		{
			name: "insert with same timestamp - sorted by ChatId",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chatA", UpdatedAt: time.Unix(200, 0), Groups: []string{"g1"}},
				{ChatId: "chatC", UpdatedAt: time.Unix(200, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chatB", Title: "SameTime", Groups: []string{"g1"}, UpdatedAt: time.Unix(200, 0)},
			},
			expectedOrder:  []string{"chatA", "chatB", "chatC"},
			mockCacheSetup: []any{nil},
		},
		{
			name: "multiple insertions maintain sort order",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chat5", UpdatedAt: time.Unix(500, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chat3", Title: "Third", Groups: []string{"g1"}, UpdatedAt: time.Unix(300, 0)},
				{Id: "chat7", Title: "Seventh", Groups: []string{"g1"}, UpdatedAt: time.Unix(700, 0)},
				{Id: "chat1", Title: "First", Groups: []string{"g1"}, UpdatedAt: time.Unix(100, 0)},
				{Id: "chat9", Title: "Ninth", Groups: []string{"g1"}, UpdatedAt: time.Unix(900, 0)},
			},
			expectedOrder:  []string{"chat9", "chat7", "chat5", "chat3", "chat1"},
			mockCacheSetup: []any{nil},
		},
		{
			name: "update existing changes position",
			initialSummaries: []*entity.ChatSummary{
				{ChatId: "chat3", UpdatedAt: time.Unix(300, 0), Groups: []string{"g1"}},
				{ChatId: "chat2", UpdatedAt: time.Unix(200, 0), Groups: []string{"g1"}},
				{ChatId: "chat1", UpdatedAt: time.Unix(100, 0), Groups: []string{"g1"}},
			},
			chatsToSave: []*entity.Chat{
				{Id: "chat1", Title: "Updated", Groups: []string{"g1"}, UpdatedAt: time.Unix(400, 0)},
			},
			expectedOrder:  []string{"chat1", "chat3", "chat2"},
			mockCacheSetup: []any{nil},
		},
		{
			name:             "empty initial summaries",
			initialSummaries: []*entity.ChatSummary{},
			chatsToSave: []*entity.Chat{
				{Id: "chat1", Title: "First", Groups: []string{"g1"}, UpdatedAt: time.Unix(100, 0)},
			},
			expectedOrder:  []string{"chat1"},
			mockCacheSetup: []any{nil},
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncCacheHit").Return().Maybe()
	mockMetricsServ.On("IncCacheMiss").Return().Maybe()
	mockMetricsServ.On("RecordCacheDuration", mock.Anything, mock.Anything).Return().Maybe()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)
			mockCache := servmocks.NewMockCache(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return(test.initialSummaries, nil)

			// Mock all Create and ValidateChat calls
			for _, chat := range test.chatsToSave {
				mockRepo.On("Create", mock.Anything, chat, mock.AnythingOfType("*entity.ChatSummary")).Return(nil)
				mockVal.On("ValidateChat", mock.Anything, chat).Return(nil)
			}

			if test.mockCacheSetup != nil {
				for _, chat := range test.chatsToSave {
					mockCache.On("Store", mock.Anything, chat).Return(test.mockCacheSetup...)
				}
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer, mockMetricsServ)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}

			// Save all chats
			for _, chat := range test.chatsToSave {
				err = svc.SaveChat(context.Background(), chat)
				assert.NoError(t, err)
			}

			// Verify order using LoadSummaries
			summaries, hasMore, err := svc.LoadSummaries(context.Background(), 0, len(test.expectedOrder)+10)
			assert.NoError(t, err)
			assert.False(t, hasMore, "Should not have more summaries than requested")
			assert.Equal(t, len(test.expectedOrder), len(summaries), "Unexpected number of summaries")

			// Verify the order matches expected
			actualOrder := make([]string, len(summaries))
			for i, summary := range summaries {
				actualOrder[i] = summary.ChatId
			}
			assert.Equal(t, test.expectedOrder, actualOrder, "Summaries not in expected order")

			// Verify summaries are properly sorted
			for i := 0; i < len(summaries)-1; i++ {
				cmpResult := summaries[i].Cmp(summaries[i+1])
				assert.LessOrEqual(t, cmpResult, 0, "Summaries at positions %d and %d are not properly sorted", i, i+1)
			}

			mockRepo.AssertExpectations(t)
			mockVal.AssertExpectations(t)
		})
	}
}

// TestChatStorageSaveChatUpdatesFields verifies that updating an existing chat
// properly updates all fields in the summary
func TestChatStorageSaveChatUpdatesFields(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	initialSummary := &entity.ChatSummary{
		ChatId:    "chat1",
		Title:     "Old Title",
		Groups:    []string{"oldGroup"},
		UpdatedAt: time.Unix(100, 0),
	}

	updatedChat := &entity.Chat{
		Id:        "chat1",
		Title:     "New Title",
		Groups:    []string{"newGroup1", "newGroup2"},
		UpdatedAt: time.Unix(200, 0),
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncCacheHit").Return().Maybe()
	mockMetricsServ.On("IncCacheMiss").Return().Maybe()
	mockMetricsServ.On("RecordCacheDuration", mock.Anything, mock.Anything).Return().Maybe()
	mockRepo := mocks.NewMockChatStorageRepository(t)
	mockVal := servmocks.NewMockValidator(t)
	mockCache := servmocks.NewMockCache(t)

	mockRepo.On("ListAll", mock.Anything).Return([]*entity.ChatSummary{initialSummary}, nil)
	mockRepo.On("Create", mock.Anything, updatedChat, mock.AnythingOfType("*entity.ChatSummary")).Return(nil)
	mockVal.On("ValidateChat", mock.Anything, updatedChat).Return(nil)
	mockCache.On("Store", mock.Anything, updatedChat).Return(nil)

	svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer, mockMetricsServ)
	assert.NoError(t, err)

	err = svc.SaveChat(context.Background(), updatedChat)
	assert.NoError(t, err)

	// Load summaries and verify the update
	summaries, _, err := svc.LoadSummaries(context.Background(), 0, 10)
	assert.NoError(t, err)
	assert.Len(t, summaries, 1)

	updatedSummary := summaries[0]
	assert.Equal(t, "chat1", updatedSummary.ChatId)
	assert.Equal(t, "New Title", updatedSummary.Title)
	assert.Equal(t, []string{"newGroup1", "newGroup2"}, updatedSummary.Groups)
	assert.Equal(t, time.Unix(200, 0), updatedSummary.UpdatedAt)

	mockRepo.AssertExpectations(t)
	mockVal.AssertExpectations(t)
}

//nolint:funlen
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

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncCacheHit").Return().Maybe()
	mockMetricsServ.On("IncCacheMiss").Return().Maybe()
	mockMetricsServ.On("RecordCacheDuration", mock.Anything, mock.Anything).Return().Maybe()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)
			mockCache := servmocks.NewMockCache(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return([]*entity.ChatSummary{}, nil).Maybe()

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
			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer, mockMetricsServ)
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
			name:           "error - limit is 0",
			offset:         0,
			limit:          0,
			initialSummary: allSummaries,
			wantErr:        true,
			ctx:            context.Background(),
			groups:         []string{},
		},
		{
			name:           "error - limit is negative",
			offset:         0,
			limit:          -5,
			initialSummary: allSummaries,
			wantErr:        true,
			ctx:            context.Background(),
			groups:         []string{},
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
			limit:          2,
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

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncCacheHit").Return().Maybe()
	mockMetricsServ.On("IncCacheMiss").Return().Maybe()
	mockMetricsServ.On("RecordCacheDuration", mock.Anything, mock.Anything).Return().Maybe()
	mockCache := servmocks.NewMockCache(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			// Mock ListAll for initialization
			mockRepo.On("ListAll", mock.Anything).Return(test.initialSummary, nil)

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, mockCache, tracer, mockMetricsServ)
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
