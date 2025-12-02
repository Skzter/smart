package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

//nolint:dupl
func TestNewValidatorService(t *testing.T) {
	service := mocks.NewMockOpenAI(t)
	logger := slog.Default()
	cfg := config.Config{}
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		service srv.OpenAI
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
			repo, err := NewValidatorService(test.service, test.config, test.logger, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewValidatorService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewValidatorService() returned nil service")
			}
		})
	}
}

//nolint:funlen
func TestValidatePrompt(t *testing.T) {
	// Setup
	serviceMock := mocks.NewMockOpenAI(t)
	cfg := &config.Config{
		Model: "gpt-4",
		Prompts: &config.Prompts{
			ValidationPrompt: "You are a helpful assistant.",
		},
	}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	svc, err := NewValidatorService(serviceMock, cfg, logger, tracer)
	if err != nil {
		t.Fatalf("failed to create validator service: %v", err)
	}

	// Define test cases
	tests := []struct {
		name      string
		request   *entity.UserRequest
		mockErr   error
		mockResp  sharedEntity.Message
		wantValid bool
		wantMsg   string
		wantErr   error
	}{
		{
			name:    "valid prompt",
			request: &entity.UserRequest{Prompt: "valid prompt"},
			mockErr: nil,
			mockResp: sharedEntity.Message{
				Body: `{
					"valid": true,
					"message": ""
				}`,
			},
			wantValid: true,
			wantMsg:   "",
			wantErr:   nil,
		},
		{
			name:    "invalid prompt",
			request: &entity.UserRequest{Prompt: "invalid prompt"},
			mockErr: nil,
			mockResp: sharedEntity.Message{
				Body: `{
					"valid": false,
					"message": "alle gründe warum es schiefgelaufen"
				}`,
			},
			wantValid: false,
			wantMsg:   "alle gründe warum es schiefgelaufen",
			wantErr:   nil,
		},
		{
			name:    "invalid JSON response",
			request: &entity.UserRequest{Prompt: "invalid json"},
			mockErr: nil,
			mockResp: sharedEntity.Message{
				Body: `invalid json`,
			},
			wantValid: false,
			wantMsg:   "",
			wantErr:   sharedErrors.ErrInternalServer,
		},
		{
			name:      "service error",
			request:   &entity.UserRequest{Prompt: "service error"},
			mockErr:   sharedErrors.ErrValidation,
			mockResp:  sharedEntity.Message{},
			wantValid: false,
			wantMsg:   "",
			wantErr:   sharedErrors.ErrValidation,
		},
		{
			name:      "invalid",
			request:   &entity.UserRequest{Prompt: "service error"},
			mockErr:   sharedErrors.ErrValidation,
			mockResp:  sharedEntity.Message{},
			wantValid: false,
			wantMsg:   "",
			wantErr:   sharedErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock.EXPECT().Request(mock.Anything, mock.Anything).
				Return(&tt.mockResp, tt.mockErr).Times(1)

			// Call ValidatePrompt
			valid, msg, err := svc.ValidatePrompt(context.Background(), &entity.Chat{}, tt.request)

			// Check results
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if valid != tt.wantValid {
					t.Fatalf("expected valid=%v, got %v", tt.wantValid, valid)
				}
				if msg != tt.wantMsg {
					t.Fatalf("expected msg=%q, got %q", tt.wantMsg, msg)
				}
			}

			// Assert that the mock was called as expected
			serviceMock.AssertExpectations(t)
		})
	}
}

func TestValidateRequest(t *testing.T) {
	serviceMock := mocks.NewMockOpenAI(t)
	cfg := &config.Config{}
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	svc, _ := NewValidatorService(serviceMock, cfg, logger, tracer)
	model := "gpt-4"
	systemPrompt := "You are a helpful assistant."
	msg := []*sharedEntity.Message{
		{
			Id:        "",
			Role:      sharedEntity.RoleUser,
			Body:      "some message",
			CreatedAt: time.Now(),
		},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		request     sharedEntity.Request
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid request",
			ctx:  context.Background(),
			request: sharedEntity.Request{
				Messages:     msg,
				Model:        model,
				SystemPrompt: systemPrompt,
			},
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "nil ctx",
			ctx:  nil,
			request: sharedEntity.Request{
				Messages:     msg,
				Model:        model,
				SystemPrompt: systemPrompt,
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrInternalServer,
		},
		{
			name: "empty Model",
			ctx:  context.Background(),
			request: sharedEntity.Request{
				Messages:     msg,
				Model:        "",
				SystemPrompt: systemPrompt,
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrValidation,
		},
		{
			name: "empty SystemPrompt",
			ctx:  context.Background(),
			request: sharedEntity.Request{
				Messages:     msg,
				Model:        model,
				SystemPrompt: "",
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrValidation,
		},
		{
			name: "empty Messages",
			ctx:  context.Background(),
			request: sharedEntity.Request{
				Messages:     []*sharedEntity.Message{},
				Model:        model,
				SystemPrompt: systemPrompt,
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateRequest(tt.ctx, tt.request)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got nil error, expected => %v", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Fatalf("unexpected error: got => %v, wanted => %v", err, tt.expectedErr)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: got => %v, wanted nil", err)
				}
			}
		})
	}
}

// nolint:funlen
func TestValidateChat(t *testing.T) {
	serviceMock := mocks.NewMockOpenAI(t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	svc, err := NewValidatorService(serviceMock, &config.Config{}, logger, tracer)
	if err != nil {
		t.Fatalf("failed to create validator service: %v", err)
	}
	id := uuid.NewString()

	tests := []struct {
		name    string
		chat    *entity.Chat
		ctx     context.Context
		wantErr bool
	}{
		{
			name: "valid Chat",
			chat: &entity.Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{Message: sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Now()}, Type: entity.MessageTypeValidation}},
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil assert failed",
			chat:    nil,
			ctx:     nil,
			wantErr: true,
		},
		{
			name: "empty id",
			chat: &entity.Chat{
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{Message: sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Now()}, Type: entity.MessageTypeValidation}},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "empty userId",
			chat: &entity.Chat{
				Id:                       "chat123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{Message: sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Now()}, Type: entity.MessageTypeValidation}},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "empty Messages",
			chat: &entity.Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "invalid message",
			chat: &entity.Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{}},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "updatedAt zero",
			chat: &entity.Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Now(),
				UpdatedAt:                time.Time{},
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{Message: sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Now()}, Type: entity.MessageTypeValidation}},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name: "createdAt zero",
			chat: &entity.Chat{
				Id:                       "chat123",
				UserId:                   "user123",
				CreatedAt:                time.Time{},
				UpdatedAt:                time.Now(),
				LastTest:                 "test123",
				LastAutoPlaywrightPrompt: "apw prompt",
				Messages:                 []entity.Message{{Message: sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Now()}, Type: entity.MessageTypeValidation}},
			},
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := svc.ValidateChat(test.ctx, test.chat)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateMessages(t *testing.T) {
	serviceMock := mocks.NewMockOpenAI(t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	svc, err := NewValidatorService(serviceMock, &config.Config{}, logger, tracer)
	if err != nil {
		t.Fatalf("failed to create validator service: %v", err)
	}
	id := uuid.NewString()
	now := time.Now()

	tests := []struct {
		name    string
		message *sharedEntity.Message
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "valid Message",
			message: &sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: now},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "nil assert failed",
			message: nil,
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "empty id",
			message: &sharedEntity.Message{Role: "role", Body: "body", CreatedAt: now},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "empty role",
			message: &sharedEntity.Message{Id: id, Body: "body", CreatedAt: now},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "empty body",
			message: &sharedEntity.Message{Id: id, Role: "role", CreatedAt: now},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "invalid createdAt",
			message: &sharedEntity.Message{Id: id, Role: "role", Body: "body", CreatedAt: time.Time{}},
			ctx:     context.Background(),
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := svc.ValidateMessage(test.ctx, test.message)

			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
