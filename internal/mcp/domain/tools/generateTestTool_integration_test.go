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

func TestGenerateTemplateTool_Integration(t *testing.T) {
	baseURL := SetupAutotesterAPIMock(t, "../../../../api/AutotesterAPI.yaml")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpClient := &http.Client{}
	tracer := otel.Tracer("test")

	repo, err := repository.NewAutotesterAPIRepository(logger, httpClient, baseURL, tracer)
	require.NoError(t, err)

	autotesterService, err := service.NewAutotesterAPIService(logger, repo)
	require.NoError(t, err)

	tool, err := NewGenerateTestTool(logger, autotesterService)
	require.NoError(t, err)

	t.Run("Successfully generate test from mock", func(t *testing.T) {
		ctx := context.Background()

		input := entity.GenerateTestRequest{
			Prompt: "generiere einen simplen Test",
			UserId: "2ed",
			ChatId: "672",
		}

		_, output, err := tool.GenerateTest(ctx, nil, input)

		require.NoError(t, err)
		//require.NotEmpty(t, output.Result.Body)
		require.NotEmpty(t, output.ChatId)
		require.NotEmpty(t, output.UserId)
	})
}
