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
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

func TestNewChat(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	storage := mocks.NewMockChatStorageService(t)
	cfg := &config.Autotester{}
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewChatManager(tt.logger, storage, cfg, tracer)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, svc)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestLoadChat(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg := &config.Autotester{
		Prompts: &config.Prompts{
			AutoPlaywrightPromptT: "ap %s",
			ValidationPrompt:      "val",
		},
	}
	tracer := otel.Tracer("test")

	now := time.Now().UTC()

	tests := []struct {
		name           string
		setupMock      func(*mocks.MockChatStorageService)
		request        entity.UserRequest
		expectErr      bool
		expectNewChat  bool
		expectedChatId string
	}{
		{
			name: "create new chat when ChatId empty",
			setupMock: func(*mocks.MockChatStorageService) {
			},
			request: entity.UserRequest{
				UserId: "user1",
				Prompt: "initial prompt",
			},
			expectErr:     false,
			expectNewChat: true,
		},
		{
			name: "load existing chat from storage",
			setupMock: func(storage *mocks.MockChatStorageService) {
				storage.On("LoadChat", mock.Anything, "chat2").Return(&entity.Chat{
					Id:             "chat2",
					Author:         "user2",
					LastModifiedBy: "user2",
				}, nil).Once()
			},
			request: entity.UserRequest{
				UserId: "user2",
				ChatId: "chat2",
			},
			expectErr:      false,
			expectNewChat:  false,
			expectedChatId: "chat2",
		},
		{
			name: "nil ctx returns error",
			setupMock: func(*mocks.MockChatStorageService) {
			},
			request: entity.UserRequest{
				UserId: "u",
				ChatId: "c",
			},
			expectErr: true,
		},
		{
			name: "error while loading chat",
			setupMock: func(storage *mocks.MockChatStorageService) {
				storage.On("LoadChat", mock.Anything, "chat2").Return(nil, errors.New("err")).Once()
			},
			request: entity.UserRequest{
				UserId: "user2",
				ChatId: "chat2",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset mock
			storage := mocks.NewMockChatStorageService(t)
			tt.setupMock(storage)

			svc, err := NewChatManager(logger, storage, cfg, tracer)
			assert.Nil(t, err)

			ctx := context.Background()
			if tt.name == "nil ctx returns error" {
				ctx = nil
			}

			got, err := svc.LoadChat(ctx, tt.request)
			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Nil(t, got)
				return
			}
			assert.Nil(t, err)
			assert.NotNil(t, got)
			if tt.expectNewChat {
				assert.NotEmpty(t, got.Id)
				assert.True(t, got.CreatedAt.After(now.Add(-time.Minute)))
				assert.Len(t, got.Messages, 0)
			} else {
				assert.Equal(t, tt.expectedChatId, got.Id)
			}
		})
	}
}

func TestSaveChat(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg := &config.Autotester{
		Prompts: &config.Prompts{
			AutoPlaywrightPromptT: "auto-prompt",
			ValidationPrompt:      "validation-prompt",
		},
	}
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		setupMock func(ch *entity.Chat, storage *mocks.MockChatStorageService)
		chat      *entity.Chat
		ctx       context.Context
		expectErr bool
	}{
		{
			name: "save updates timestamps and prompts",
			setupMock: func(ch *entity.Chat, storage *mocks.MockChatStorageService) {
				storage.On("SaveChat", mock.Anything, mock.AnythingOfType("*entity.Chat")).Return(nil).Once()
			},
			chat: &entity.Chat{
				Id:             "c1",
				Author:         "u1",
				LastModifiedBy: "u1",
				CreatedAt:      time.Now().Add(-time.Hour).UTC(),
			},
			ctx:       context.Background(),
			expectErr: false,
		},
		{
			name: "nil ctx returns error and does not call storage",
			setupMock: func(ch *entity.Chat, storage *mocks.MockChatStorageService) {
			},
			chat: &entity.Chat{
				Id:             "c2",
				Author:         "u2",
				LastModifiedBy: "u2",
			},
			ctx:       nil,
			expectErr: true,
		},
		{
			name: "save chat error",
			setupMock: func(ch *entity.Chat, storage *mocks.MockChatStorageService) {
				storage.On("SaveChat", mock.Anything, mock.AnythingOfType("*entity.Chat")).Return(errors.New("err")).Once()
			},
			chat: &entity.Chat{
				Id:             "c1",
				Author:         "u1",
				LastModifiedBy: "u1",
				CreatedAt:      time.Now().Add(-time.Hour).UTC(),
			},
			ctx:       context.Background(),
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset mock
			storage := mocks.NewMockChatStorageService(t)
			tt.setupMock(tt.chat, storage)

			svc, err := NewChatManager(logger, storage, cfg, tracer)
			assert.Nil(t, err)

			err = svc.SaveChat(tt.ctx, tt.chat, "user")
			if tt.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			// verify updated fields
			assert.False(t, tt.chat.UpdatedAt.IsZero())
		})
	}
}
