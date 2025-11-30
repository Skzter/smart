package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

// nolint: dupl
func TestNewGeneratePromptService(t *testing.T) {
	openai := mocks.NewMockOpenAI(t)
	taglist := mocks.NewMockTaglistStorage(t)
	logger := slog.Default()
	cfg := config.Config{}

	tests := []struct {
		name    string
		openai  srv.OpenAI
		taglist srv.TaglistStorage
		config  *config.Config
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			openai:  openai,
			taglist: taglist,
			config:  &cfg,
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil openai",
			openai:  nil,
			taglist: taglist,
			config:  &cfg,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil config",
			openai:  openai,
			taglist: taglist,
			config:  nil,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil logger",
			openai:  openai,
			taglist: taglist,
			config:  &cfg,
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewGeneratePromptService(test.openai, test.taglist, test.config, test.logger)
			if test.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, repo)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

//nolint:funlen
func TestGeneratePrompt(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		Model: "gpt-4",
		Prompts: &config.Prompts{
			AutoPlaywrightPromptT: "system prompt %s",
		},
	}
	generatedCode := "some code"
	tags := &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "Tag1", Description: ""}, {Name: "Tag2", Description: ""}}}

	tests := []struct {
		name           string
		setup          func(*mocks.MockOpenAI, *mocks.MockTaglistStorage)
		ctx            context.Context
		chat           *entity.Chat
		request        *entity.UserRequest
		expectedResult string
		wantErr        bool
	}{
		{
			name: "successful generation",
			setup: func(openai *mocks.MockOpenAI, taglist *mocks.MockTaglistStorage) {
				taglist.EXPECT().GetTaglist(mock.Anything).
					Return(tags, nil)
				openai.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Role: "assistant", Body: generatedCode}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user123", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat123",
				Prompt: "Generate test for login form",
				UserId: "user123",
			},
			expectedResult: generatedCode,
			wantErr:        false,
		},
		{
			name: "empty response body",
			setup: func(openai *mocks.MockOpenAI, taglist *mocks.MockTaglistStorage) {
				taglist.EXPECT().GetTaglist(mock.Anything).
					Return(tags, nil)
				openai.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Role: "assistant", Body: ""}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user456", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat456",
				Prompt: "Generate test",
				UserId: "user456",
			},
			expectedResult: "",
			wantErr:        true,
		},
		{
			name: "openai service error",
			setup: func(openai *mocks.MockOpenAI, taglist *mocks.MockTaglistStorage) {
				taglist.EXPECT().GetTaglist(mock.Anything).
					Return(tags, nil)
				openai.EXPECT().Request(mock.Anything, mock.Anything).
					Return(nil, sharedErrors.ErrInternalServer)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user789", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat789",
				Prompt: "Generate test",
				UserId: "user789",
			},
			expectedResult: "",
			wantErr:        true,
		},
		{
			name: "taglist fetch error",
			setup: func(openai *mocks.MockOpenAI, taglist *mocks.MockTaglistStorage) {
				taglist.EXPECT().GetTaglist(mock.Anything).
					Return(nil, sharedErrors.ErrInternalServer)
				openai.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Role: "assistant", Body: generatedCode}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user101", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat101",
				Prompt: "Generate test",
				UserId: "user101",
			},
			expectedResult: generatedCode,
			wantErr:        false,
		},
		{
			name: "nil context",
			setup: func(openai *mocks.MockOpenAI, taglist *mocks.MockTaglistStorage) {
				// No mock expectations needed as function should fail before any calls
			},
			ctx:  nil,
			chat: entity.NewChat("user202", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat202",
				Prompt: "Generate test",
				UserId: "user202",
			},
			expectedResult: "",
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockOpenAI := mocks.NewMockOpenAI(t)
			mockTaglist := mocks.NewMockTaglistStorage(t)
			test.setup(mockOpenAI, mockTaglist)

			svc, err := NewGeneratePromptService(mockOpenAI, mockTaglist, cfg, logger)
			assert.Nil(t, err)
			assert.NotNil(t, svc)

			result, err := svc.GeneratePrompt(test.ctx, test.chat, test.request)

			if test.wantErr {
				assert.NotNil(t, err)
				assert.Empty(t, result)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}
