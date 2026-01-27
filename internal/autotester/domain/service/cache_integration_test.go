//go:build integration

package service

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcValkey "github.com/testcontainers/testcontainers-go/modules/valkey"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"go.opentelemetry.io/otel"
)

func TestCache(t *testing.T) {
	ctx := t.Context()
	logger := slog.New(slog.DiscardHandler)

	valkeyTestContainer, err := tcValkey.Run(ctx, "valkey/valkey:alpine")
	require.NoError(t, err)
	t.Cleanup(func() {
		err := testcontainers.TerminateContainer(valkeyTestContainer)
		require.NoError(t, err, "error when terminating valkey testcontainer")
	})

	host, err := valkeyTestContainer.Host(ctx)
	require.NoError(t, err)

	port, err := valkeyTestContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	autotesterConfig, err := config.LoadConfig()
	require.NoError(t, err)

	autotesterConfig.RedisConfig.Addr = fmt.Sprintf("%s:%d", host, port.Int())
	autotesterConfig.RedisConfig.Password = ""
	autotesterConfig.RedisConfig.Db = 0
	autotesterConfig.RedisConfig.Protocol = 2

	cacheRepo, err := repository.NewRedisCache(logger, autotesterConfig.RedisConfig)
	require.NoError(t, err)

	tracer := otel.Tracer("integration test")
	cacheServ, err := NewCacheService(autotesterConfig, logger, cacheRepo, tracer)
	require.NoError(t, err)

	data := []*entity.Chat{
		entity.NewChat("user1", nil),
		entity.NewChat("user2", nil),
		entity.NewChat("user3", nil),
		entity.NewChat("user4", nil),
		entity.NewChat("user5", nil),
	}

	t.Logf("storing data into the cache")
	for _, chat := range data {
		err := cacheServ.Store(ctx, chat)
		require.NoError(t, err)
	}

	// cache hit
	t.Logf("looking up chat with id %s (cache hit)\n", data[0].Id)
	cachedChat1, err := cacheServ.LookUp(ctx, data[0].Id)
	require.NoError(t, err)
	require.NotNil(t, cachedChat1, "expected cache hit")
	require.Equal(t, cachedChat1.Id, data[0].Id, "expected same chat (same ids)")
	t.Logf("hit")

	// cache miss
	t.Logf("looking up chat with id: user6 (cache miss)\n")
	cachedChat6, err := cacheServ.LookUp(ctx, "user6")
	require.NoError(t, err)
	require.Nil(t, cachedChat6, "expected cache miss -> nil value in cachedChat6")
	t.Logf("miss")
	t.Logf("integration test worked successfully")
}
