//go:build integration

package service

import (
	"log/slog"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"go.opentelemetry.io/otel"
)

func TestCache(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)
	autotesterConfig, err := config.LoadConfig()
	require.NoError(t, err)
	cacheRepo, err := repository.NewRedisCache(logger, autotesterConfig.RedisConfig)
	require.NoError(t, err)
	tracer := otel.Tracer("integration test")
	cacheServ, err := NewCacheService(autotesterConfig, logger, cacheRepo, tracer)
	require.NoError(t, err)

	cmd := exec.CommandContext(ctx, "docker-compose",
		"-f", "../../../../deployments/compose.dev.yml",
		"up", "-d", "redis", "--wait")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to start docker-compose: %v", err)
	}

	t.Cleanup(func() {
		cleanupCmd := exec.CommandContext(ctx, "docker-compose",
			"-f", "../../../../deployments/compose.dev.yml",
			"down")
		cleanupCmd.Run()
	})

	time.Sleep(2 * time.Second)

	data := []*entity.Chat{
		entity.NewChat("user1", nil),
		entity.NewChat("user2", nil),
		entity.NewChat("user3", nil),
		entity.NewChat("user4", nil),
		entity.NewChat("user5", nil),
	}

	for _, chat := range data {
		err := cacheServ.Store(ctx, chat)
		require.NoError(t, err)
	}

	// cache hit
	cachedChat1, err := cacheServ.LookUp(ctx, data[0].Id)
	require.NoError(t, err)
	require.NotNil(t, cachedChat1, "expected cache hit")
	require.Equal(t, cachedChat1.Id, data[0].Id, "expected same chat (same ids)")

	// cache miss
	cachedChat6, err := cacheServ.LookUp(ctx, "user6")
	require.NoError(t, err)
	require.Nil(t, cachedChat6, "expected cache miss -> nil value in cachedChat6")
}
