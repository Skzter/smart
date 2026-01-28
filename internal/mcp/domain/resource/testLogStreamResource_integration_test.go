//go:build integration

package resource

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
)

func TestTestLogStreamResource_Integration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name  string
		input *mcp.ReadResourceRequest
		setup func(store.TestLogStreamStore)
		check func(*testing.T, *mcp.ReadResourceResult, error)
	}{
		{
			name: "Error when reading non-existent log stream",
			input: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/test-empty/logs",
				},
			},
			setup: nil,
			check: func(t *testing.T, result *mcp.ReadResourceResult, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "Successfully read completed log stream",
			input: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/test-completed/logs",
				},
			},
			setup: func(s store.TestLogStreamStore) {
				s.AddEvent("test-completed", entity.LogEvent{
					Event: "progress",
					Data:  `{"status":"completed"}`,
				})
				s.CompleteStream("test-completed")
			},
			check: func(t *testing.T, result *mcp.ReadResourceResult, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)

				var payload map[string]any
				err = json.Unmarshal([]byte(result.Contents[0].Text), &payload)
				require.NoError(t, err)

				require.True(t, payload["meta"].(map[string]any)["final"].(bool))
			},
		},
		{
			name: "Error with invalid URI format",
			input: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "invalid://format",
				},
			},
			check: func(t *testing.T, result *mcp.ReadResourceResult, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
		{
			name: "Error with non-existent stream",
			input: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/non-existent/logs",
				},
			},
			check: func(t *testing.T, result *mcp.ReadResourceResult, err error) {
				require.Error(t, err)
				require.Nil(t, result)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testStore := store.NewTestLogStreamStore()
			defer testStore.Shutdown()

			resource, err := NewTestLogStreamResource(logger, testStore)
			require.NoError(t, err)

			if test.setup != nil {
				test.setup(testStore)
			}

			ctx := context.Background()
			result, err := resource.ReadTestLogStream(ctx, test.input)

			test.check(t, result, err)
		})
	}
}
