package service

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// nolint: dupl
func TestNewGeneratePromptService(t *testing.T) {
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
			repo, err := NewGeneratePromptService(test.service, test.config, test.logger)
			if (err != nil) != test.wantErr {
				t.Errorf("NewGeneratePromptService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewGeneratePromptService() returned nil service")
			}
		})
	}
}

func TestGeneratePrompt(t *testing.T) {
	logger := slog.Default()
	cfg := config.Config{
		Prompts: &config.Prompts{
			AutoPlaywrightPrompt: "system prompt",
		},
	}

	tests := []struct {
		name     string
		wantText string
		wantErr  bool
	}{
		{
			name:     "success",
			wantText: "openai says hello",
			wantErr:  false,
		},
		{
			name:    "service error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := mocks.NewMockOpenAIService(t)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			if tt.wantErr {
				service.
					On("Request", ctx, mock.Anything).
					Return((*entity.Response)(nil), fmt.Errorf("service failure"))
			} else {
				service.
					On("Request", ctx, mock.Anything).
					Return(&entity.Response{Text: tt.wantText}, nil)
			}

			svc := &GeneratePromptService{
				service: service,
				config:  &cfg,
				logger:  logger,
			}
			got, err := svc.GeneratePrompt(ctx, "user says hi", "session-123")
			if (err != nil) != tt.wantErr {
				t.Fatalf("GeneratePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.wantText {
				t.Errorf("GeneratePrompt() = %q, want %q", got, tt.wantText)
			}

			service.AssertExpectations(t)
		})
	}
}
