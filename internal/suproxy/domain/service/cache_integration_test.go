//go:build integration
package service

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tcRedis "github.com/testcontainers/testcontainers-go/modules/redis"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// newIntegrationLogger returns a logger that discards all output
func newIntegrationLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestCacheService_Integration_WithRedis verifies end-to-end cache behavior using a Redis Testcontainers module instance
func TestCacheService_Integration_WithRedis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Start Redis via testcontainers module
	redisC, err := tcRedis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	defer func() { _ = redisC.Terminate(ctx) }()

	rawAddr, err := redisC.ConnectionString(ctx)
	require.NoError(t, err)

	// Remove "redis://" prefix → go-redis needs plain "host:port"
	addr := strings.TrimPrefix(rawAddr, "redis://")

	// Create Redis config pointing to the test container
	logger := newIntegrationLogger()
	cfg := &config.Config{
		Redis: &config.RedisConfig{
			Addr:     addr,
			Password: "",
			Db:       0,
			Protocol: 2,
		},
	}

	// Initialize real Redis cache repository
	cacheRepo, err := repository.NewRedisCache(logger, cfg)
	require.NoError(t, err)
	defer func() {
		_ = cacheRepo.Close() // ignore close error intentionally
	}()

	// Create CacheService instance using the real repository
	cacheSvc := NewCacheService(logger, cfg, cacheRepo)

	// Build example request used as cache key input
	reqEntity := entity.Request{
		Destination: "/integration-test",
		Header: map[string]string{
			"X-Test": "1",
		},
		Request: `{"foo":"bar"}`,
	}

	originalResponse := []byte(`{"result":"ok"}`)

	// Store the response in Redis via the CacheService
	err = cacheSvc.Store(ctx, reqEntity, originalResponse, false, false)
	require.NoError(t, err)

	// Short wait to ensure Redis write is processed
	time.Sleep(50 * time.Millisecond)

	// Retrieve cached value via Lookup
	cachedResponse, hit, err := cacheSvc.Lookup(ctx, reqEntity, false)
	require.NoError(t, err)

	// Assert cache hit and correct value returned
	require.True(t, hit, "expected cache hit")
	require.Equal(t, string(originalResponse), string(cachedResponse))
}
