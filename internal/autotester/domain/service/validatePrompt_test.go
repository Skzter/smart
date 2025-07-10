package service

import (
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

//nolint:dupl
func TestNewValidatePromptService(t *testing.T) {
	service := &mocks.MockOpenAIService{}
	logger := slog.Default()
	cfg := config.Config{}

	tests := []struct {
		name    string
		service srv.OpenAIService
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
