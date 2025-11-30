package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

//nolint:dupl
func TestNewValidatorService(t *testing.T) {
	service := mocks.NewMockOpenAI(t)
	logger := slog.Default()
	cfg := config.Config{}

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
			repo, err := NewValidatorService(test.service, test.config, test.logger)
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
	svc, err := NewValidatorService(serviceMock, cfg, logger)
	if err != nil {
		t.Fatalf("failed to create validator service: %v", err)
	}

	// Define test cases
	tests := []struct {
		name       string
		userPrompt string
		mockErr    error
		mockResp   entity.Message
		wantValid  bool
		wantMsg    string
		wantErr    error
	}{
		{
			name:       "valid prompt",
			userPrompt: "valid Prompt",
			mockErr:    nil,
			mockResp: entity.Message{
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
			name:       "invalid prompt",
			userPrompt: "invalid Prompt",
			mockErr:    nil,
			mockResp: entity.Message{
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
			name:       "invalid JSON response",
			userPrompt: "invalid json",
			mockErr:    nil,
			mockResp: entity.Message{
				Body: `invalid json`,
			},
			wantValid: false,
			wantMsg:   "",
			wantErr:   sharedErrors.ErrInternalServer,
		},
		{
			name:       "service error",
			userPrompt: "service error",
			mockErr:    sharedErrors.ErrValidation,
			mockResp:   entity.Message{},
			wantValid:  false,
			wantMsg:    "",
			wantErr:    sharedErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock.EXPECT().Request(mock.Anything, mock.Anything).
				Return(&tt.mockResp, tt.mockErr).Times(1)

			// Call ValidatePrompt
			valid, msg, err := svc.ValidatePrompt(context.Background(), tt.userPrompt)

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
	svc, _ := NewValidatorService(serviceMock, cfg, logger)

	model := "gpt-4"
	systemPrompt := "You are a helpful assistant."
	msg := []entity.Message{
		{
			Id:        "",
			Role:      entity.RoleUser,
			Body:      "some message",
			CreatedAt: time.Now(),
		},
	}

	tests := []struct {
		name        string
		ctx         context.Context
		request     entity.Request
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid request",
			ctx:  context.Background(),
			request: entity.Request{
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
			request: entity.Request{
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
			request: entity.Request{
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
			request: entity.Request{
				Messages:     msg,
				Model:        model,
				SystemPrompt: "",
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
