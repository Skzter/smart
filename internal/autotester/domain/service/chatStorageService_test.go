package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func validChat() *entity.Chat {
	return &entity.Chat{
		Id:                       "chat123",
		UserId:                   "user123",
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
		LastTest:                 "test123",
		LastAutoPlaywrightPrompt: "apw prompt",
		Messages:                 []entity.Message{{Type: entity.TypeAny, Message: sharedEntity.Message{Id: "id", Role: "user", Body: "msg"}}},
	}
}

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockChatStorageRepository(t)

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.ChatStorageRepository
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			repo:    mockRepo,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			repo:    mockRepo,
			wantErr: true,
		},
		{
			name:    "nil repo",
			logger:  logger,
			repo:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewChatStorageService(test.logger, test.repo)
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
	logger := slog.Default()

	tests := []struct {
		name          string
		createReturns []any
		wantErr       bool
		chat          *entity.Chat
	}{
		{
			name:          "success",
			createReturns: []any{nil},
			wantErr:       false,
			chat:          validChat(),
		},
		{
			name:    "nil chat",
			wantErr: true,
			chat:    nil,
		},
		{
			name:    "validation error",
			wantErr: true,
			chat:    &entity.Chat{},
		},
		{
			name:          "repo returns error",
			createReturns: []any{errors.New("repo error")},
			wantErr:       true,
			chat:          validChat(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.chat).Return(test.createReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
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

func TestChatStorageLoadChat(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		userid      string
		chatid      string
		loadReturns []any
		wantErr     bool
	}{
		{
			name:        "success",
			userid:      "user123",
			chatid:      "chat123",
			loadReturns: []any{validChat(), nil},
			wantErr:     false,
		},
		{
			name:    "invalid userId",
			userid:  "",
			chatid:  "chat123",
			wantErr: true,
		},
		{
			name:    "invalid chatId",
			userid:  "user123",
			chatid:  "",
			wantErr: true,
		},
		{
			name:        "repo returns error",
			userid:      "user123",
			chatid:      "chat123",
			loadReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
		},
		{
			name:        "repo returns invalid chat",
			userid:      "user123",
			chatid:      "chat123",
			loadReturns: []any{&entity.Chat{}, nil},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.loadReturns != nil {
				mockRepo.On("Read", mock.Anything, test.userid, test.chatid).Return(test.loadReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadChat(context.Background(), test.userid, test.chatid)
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

func TestLoadUserChats(t *testing.T) {
	logger := slog.Default()
	orderedResult := []*entity.ChatSummary{{UpdatedAt: time.Unix(200, 0)}, {UpdatedAt: time.Unix(100, 0)}}

	tests := []struct {
		name        string
		userId      string
		listReturns []any
		wantErr     bool
	}{
		{
			name:        "success",
			userId:      "user123",
			listReturns: []any{orderedResult, nil},
			wantErr:     false,
		},
		{
			name:        "inverse order",
			userId:      "user123",
			listReturns: []any{[]*entity.ChatSummary{{UpdatedAt: time.Unix(100, 0)}, {UpdatedAt: time.Unix(200, 0)}}, nil},
			wantErr:     false,
		},
		{
			name:    "invalid userId",
			userId:  "",
			wantErr: true,
		},
		{
			name:        "repo returns error",
			userId:      "user123",
			listReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.listReturns != nil {
				mockRepo.On("FindByUserID", mock.Anything, test.userId).Return(test.listReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadUserChats(context.Background(), test.userId)
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
