package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

//nolint:dupl
func TestNewValidatePromptService(t *testing.T) {
	service := mocks.NewMockOpenAIService(t)
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

func TestValidatePrompt(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		Prompts: &config.Prompts{
			ValidationPrompt: "validate system prompt",
		},
	}

	tests := []struct {
		name         string
		mockSetup    func(s *mocks.MockOpenAI, ctx context.Context)
		wantErr      bool
		errSubstring string
	}{
		{
			name: "service error",
			mockSetup: func(s *mocks.MockOpenAI, ctx context.Context) {
				s.On("Request", ctx, mock.Anything).
					Return((*entity.Response)(nil), fmt.Errorf("network down"))
			},
			wantErr:      true,
			errSubstring: "failed to send prompt validation request to OpenAI",
		},
		{
			name: "valid prompt (true)",
			mockSetup: func(s *mocks.MockOpenAI, ctx context.Context) {
				s.On("Request", ctx, mock.Anything).
					Return(&entity.Response{Text: "true"}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid prompt (false)",
			mockSetup: func(s *mocks.MockOpenAI, ctx context.Context) {
				s.On("Request", ctx, mock.Anything).
					Return(&entity.Response{Text: "false"}, nil)
			},
			wantErr:      true,
			errSubstring: "prompt does not contain required information",
		},
		{
			name: "unexpected response",
			mockSetup: func(s *mocks.MockOpenAI, ctx context.Context) {
				s.On("Request", ctx, mock.Anything).
					Return(&entity.Response{Text: "oops"}, nil)
			},
			wantErr:      true,
			errSubstring: `unexpected validation response: "oops"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewMockOpenAIService(t)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			tt.mockSetup((*mocks.MockOpenAI)(serviceMock), ctx)

			svc := &validatePrompt{
				service: serviceMock,
				config:  cfg,
				logger:  logger,
			}
			err := svc.ValidatePrompt(ctx, "some user prompt", "session-xyz")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstring) {
					t.Fatalf("error %q does not contain expected substring %q", err.Error(), tt.errSubstring)
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			serviceMock.AssertExpectations(t)
		})
	}
}
