//go:build integration

package tools

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
)

func TestRunTestTool(t *testing.T) {
	baseURL := SetupAutotesterAPIMock(t, "../../../../api/AutotesterAPI.yaml")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpClient := &http.Client{}
	tracer := otel.Tracer("test")

	repo, err := repository.NewAutotesterAPIRepository(logger, httpClient, baseURL, tracer)
	require.NoError(t, err)

	autotesterService, err := service.NewAutotesterAPIService(logger, repo)
	require.NoError(t, err)

	tool, err := NewRunTestTool(logger, autotesterService)
	require.NoError(t, err)

	t.Run("Successfully run test from mock", func(t *testing.T) {
		ctx := context.Background()

		_, output, err := tool.RunTest(ctx, nil, entity.ExecuteTestRequest{})

		require.NoError(t, err)
		require.NotEmpty(t, output.Result)
	})
}
