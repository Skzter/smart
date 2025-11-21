package service

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

func TestNewDocker(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tests := []struct {
		name    string
		logger  *slog.Logger
		config  *config.Config
		wantErr bool
	}{
		{
			name:    "success - valid docker service",
			logger:  logger,
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "error - nil values",
			logger:  nil,
			config:  nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serv, err := NewDocker(tt.logger, tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, serv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, serv)
			}
		})
	}
}
