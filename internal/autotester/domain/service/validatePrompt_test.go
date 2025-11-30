package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	testifyAssert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

//nolint:dupl
func TestNewValidatePromptService(t *testing.T) {
	service := mocks.NewMockOpenAI(t)
	logger := slog.Default()
	cfg := config.Config{}

	tests := []struct {
		name    string
		service sharedService.OpenAI
		config  *config.Config
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			service: service,
			config:  &cfg,
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil service",
			service: nil,
			config:  &cfg,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil config",
			service: service,
			config:  nil,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil logger",
			service: service,
			config:  &cfg,
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewValidatePromptService(test.service, test.config, test.logger)
			if test.wantErr {
				testifyAssert.Nil(t, repo)
				testifyAssert.NotNil(t, err)
			}
			if !test.wantErr && repo == nil {
				testifyAssert.NotNil(t, repo)
				testifyAssert.Nil(t, err)
			}
		})
	}
}

//nolint:funlen
func TestValidatePrompt(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		Model: "gpt-4",
		Prompts: &config.Prompts{
			ValidationPrompt: "Validate this prompt",
		},
	}

	tests := []struct {
		name          string
		setup         func(*mocks.MockOpenAI)
		ctx           context.Context
		chat          *entity.Chat
		request       *entity.UserRequest
		expectedValid bool
		expectedMsg   string
		wantErr       bool
	}{
		{
			name: "valid prompt",
			setup: func(m *mocks.MockOpenAI) {
				validResponse := entity.LlmValidationResponse{
					Valid:   true,
					Message: "Prompt is valid",
				}
				body, _ := json.Marshal(validResponse)
				m.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Body: string(body)}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user123", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat123",
				Prompt: "Test this login form",
				UserId: "user123",
			},
			expectedValid: true,
			expectedMsg:   "Prompt is valid",
			wantErr:       false,
		},
		{
			name: "invalid prompt",
			setup: func(m *mocks.MockOpenAI) {
				invalidResponse := entity.LlmValidationResponse{
					Valid:   false,
					Message: "Prompt is missing required information",
				}
				body, _ := json.Marshal(invalidResponse)
				m.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Body: string(body)}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user456", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat456",
				Prompt: "Test something",
				UserId: "user456",
			},
			expectedValid: false,
			expectedMsg:   "Prompt is missing required information",
			wantErr:       false,
		},
		{
			name: "service request error",
			setup: func(m *mocks.MockOpenAI) {
				m.EXPECT().Request(mock.Anything, mock.Anything).
					Return(nil, errors.New("service error"))
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user789", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat789",
				Prompt: "Test case",
				UserId: "user789",
			},
			expectedValid: false,
			expectedMsg:   "",
			wantErr:       true,
		},
		{
			name: "invalid json response",
			setup: func(m *mocks.MockOpenAI) {
				m.EXPECT().Request(mock.Anything, mock.Anything).
					Return(&sharedEntity.Message{Body: "invalid json"}, nil)
			},
			ctx:  context.Background(),
			chat: entity.NewChat("user101", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat101",
				Prompt: "Test invalid response",
				UserId: "user101",
			},
			expectedValid: false,
			expectedMsg:   "",
			wantErr:       true,
		},
		{
			name: "nil context",
			setup: func(m *mocks.MockOpenAI) {
				// No mock expectations needed as function should fail before any calls
			},
			ctx:  nil,
			chat: entity.NewChat("user202", []entity.Message{}),
			request: &entity.UserRequest{
				ChatId: "chat202",
				Prompt: "Test nil context",
				UserId: "user202",
			},
			expectedValid: false,
			expectedMsg:   "",
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := mocks.NewMockOpenAI(t)
			test.setup(mockService)

			svc, err := NewValidatePromptService(mockService, cfg, logger)
			testifyAssert.Nil(t, err)
			testifyAssert.NotNil(t, svc)

			isValid, msg, err := svc.ValidatePrompt(test.ctx, test.chat, test.request)

			if test.wantErr {
				testifyAssert.NotNil(t, err)
			} else {
				testifyAssert.Nil(t, err)
				testifyAssert.Equal(t, test.expectedValid, isValid)
				testifyAssert.Equal(t, test.expectedMsg, msg)
			}
		})
	}
}
