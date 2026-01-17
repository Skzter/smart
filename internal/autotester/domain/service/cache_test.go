package service

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

func TestNewCacheService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tracer := otel.Tracer("test")
	tests := []struct {
		name      string
		logger    *slog.Logger
		config    *config.Autotester
		cacheRepo repository.Cache
		wantError bool
	}{
		{
			name:      "success - service gets build",
			logger:    logger,
			config:    cfg,
			cacheRepo: mocks.NewMockCache(t),
			wantError: false,
		},
		{
			name:      "error - nil params",
			logger:    nil,
			config:    nil,
			cacheRepo: nil,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewCacheService(tc.config, tc.logger, tc.cacheRepo, tracer)
			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, srv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, srv)
			}
		})
	}
}
