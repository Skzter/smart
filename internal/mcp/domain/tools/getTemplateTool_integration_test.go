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
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
)

func TestGetTemplateTool_Integration(t *testing.T) {
	baseURL := SetupAutotesterAPIMock(t, "../../../../api/AutotesterAPI.yaml")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	httpClient := &http.Client{}
	tracer := otel.Tracer("test")

	repo, err := repository.NewAutotesterAPIRepository(logger, httpClient, baseURL, tracer)
	require.NoError(t, err)

	store := store.NewTestLogStreamStore()

	autotesterService, err := service.NewAutotesterAPIService(logger, repo, store)
	require.NoError(t, err)

	tool, err := NewGetTemplateTool(logger, autotesterService)
	require.NoError(t, err)

	t.Run("Successfully retrieve template from mock", func(t *testing.T) {
		ctx := context.Background()

		_, output, err := tool.GetTemplate(ctx, nil, entity.TemplateIn{})

		require.NoError(t, err)
		require.NotEmpty(t, output.Content)
	})
}
