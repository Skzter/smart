package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

//nolint:dupl
func TestNewValidatePromptService(t *testing.T) {
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
			repo, err := NewValidatePromptService(test.service, test.config, test.logger)
			if (err != nil) != test.wantErr {
				t.Errorf("NewValidatePromptService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewValidatePromptService() returned nil service")
			}
		})
	}
}

//nolint:funlen
func TestValidatePrompt(t *testing.T) {
	serviceMock := mocks.NewMockOpenAI(t)
	cfg := &config.Config{
		Model: "",
		Prompts: &config.Prompts{
			ValidationPrompt: "",
		},
	}
	logger := slog.New(slog.DiscardHandler)
	svc, _ := NewValidatePromptService(serviceMock, cfg, logger)

	sessionId := "die tollste Session ID"

	validPrompt := "valid Prompt"
	invalidPrompt := "invalid Prompt"
	invalidJson := "invalid json"

	validRequest := entity.Request{
		Prompt:    validPrompt,
		SessionID: sessionId,
	}
	validResponse := entity.Response{
		Text: `{"valid": true, "message": ""}`,
	}

	invalidRequest := entity.Request{
		Prompt:    invalidPrompt,
		SessionID: sessionId,
	}
	invalidResponse := entity.Response{
		Text: `{"valid": false, "message": "alle gründe warum es schiefgelaufen"}`,
	}

	invalidJsonRequest := entity.Request{
		Prompt:    invalidJson,
		SessionID: sessionId,
	}
	invalidJsonResponse := entity.Response{
		Text: "no json",
	}

	mockSetup := []struct {
		request     entity.Request
		response    *entity.Response
		returnError error
		wantErr     bool
	}{
		{
			request:     validRequest,
			response:    &validResponse,
			returnError: nil,
		},
		{
			request:     invalidRequest,
			response:    &invalidResponse,
			returnError: nil,
		},
		{
			request:     invalidJsonRequest,
			response:    &invalidJsonResponse,
			returnError: nil,
		},
		{
			request: entity.Request{
				Prompt:    "",
				SessionID: sessionId,
			},
			response:    &entity.Response{},
			returnError: sharedErrors.ErrInternalServer,
		},
	}

	tests := []struct {
		name        string
		userPrompt  string
		ctx         context.Context
		wantErr     bool
		isValid     bool
		expectedErr error
	}{
		{
			name:        "valid request without changes",
			userPrompt:  validPrompt,
			ctx:         context.Background(),
			wantErr:     false,
			isValid:     true,
			expectedErr: nil,
		},
		{
			name:        "valid request but with changes",
			userPrompt:  invalidPrompt,
			ctx:         context.Background(),
			wantErr:     false,
			isValid:     false,
			expectedErr: nil,
		},
		{
			name:        "invalid json",
			userPrompt:  invalidJson,
			ctx:         context.Background(),
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrValidation,
		},
		{
			name:        "nil ctx",
			userPrompt:  validPrompt,
			ctx:         nil,
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrValidation,
		},
		{
			name:        "service error",
			userPrompt:  "",
			ctx:         context.Background(),
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrValidation,
		},
	}

	for _, mc := range mockSetup {
		serviceMock.On("Request", mock.Anything, mc.request).Return(mc.response, mc.returnError)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, err := svc.ValidatePrompt(tt.ctx, tt.userPrompt, sessionId)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("got nil error, expected => %v", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					if !strings.Contains(err.Error(), tt.expectedErr.Error()) {
						t.Fatalf("unexpected error: got => %v, wanted => %v", err, tt.expectedErr)
					}
				}
			} else {
				if tt.isValid {
					if str != "" {
						t.Fatalf("expected nil string, got => %v", str)
					}
				} else {
					if str == "" {
						t.Fatal("expected populated string but got nil string")
					}
				}
			}
		})
	}
	mock.AssertExpectationsForObjects(t)
}
