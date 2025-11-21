package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

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

	sessionID := "die tollste Session ID"

	// Define test cases
	tests := []struct {
		name       string
		userPrompt string
		mockResp   entity.Response
		mockErr    error
		wantValid  bool
		wantMsg    string
		wantErr    error
	}{
		{
			name:       "valid prompt",
			userPrompt: "valid Prompt",
			mockResp:   entity.Response{Text: `{"valid": true, "message": ""}`},
			mockErr:    nil,
			wantValid:  true,
			wantMsg:    "",
			wantErr:    nil,
		},
		{
			name:       "invalid prompt",
			userPrompt: "invalid Prompt",
			mockResp:   entity.Response{Text: `{"valid": false, "message": "alle gründe warum es schiefgelaufen"}`},
			mockErr:    nil,
			wantValid:  false,
			wantMsg:    "alle gründe warum es schiefgelaufen",
			wantErr:    nil,
		},
		{
			name:       "invalid JSON response",
			userPrompt: "invalid json",
			mockResp:   entity.Response{Text: `not a json`},
			mockErr:    nil,
			wantValid:  false,
			wantMsg:    "",
			wantErr:    sharedErrors.ErrInternalServer,
		},
		{
			name:       "service error",
			userPrompt: "service error",
			mockResp:   entity.Response{},
			mockErr:    sharedErrors.ErrValidation,
			wantValid:  false,
			wantMsg:    "",
			wantErr:    sharedErrors.ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock for this test case
			req := entity.Request{
				Prompt:       tt.userPrompt,
				SessionID:    sessionID,
				Model:        cfg.Model,
				SystemPrompt: cfg.Prompts.ValidationPrompt,
			}
			serviceMock.On("Request", mock.Anything, req).Return(&tt.mockResp, tt.mockErr).Once()

			// Call ValidatePrompt
			valid, msg, err := svc.ValidatePrompt(context.Background(), tt.userPrompt, sessionID)

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

	sessionId := "die tollste Session ID"
	model := "gpt-4"
	systemPrompt := "You are a helpful assistant."

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
				SessionID:    sessionId,
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
				SessionID:    sessionId,
				Model:        model,
				SystemPrompt: systemPrompt,
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrInternalServer,
		},
		{
			name: "empty SessionID",
			ctx:  context.Background(),
			request: entity.Request{
				SessionID:    "",
				Model:        model,
				SystemPrompt: systemPrompt,
			},
			wantErr:     true,
			expectedErr: sharedErrors.ErrValidation,
		},
		{
			name: "empty Model",
			ctx:  context.Background(),
			request: entity.Request{
				SessionID:    sessionId,
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
				SessionID:    sessionId,
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
