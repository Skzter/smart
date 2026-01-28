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
	testStore := store.NewTestLogStreamStore()
	defer testStore.Shutdown()

	// Create the resource
	resource, err := NewTestLogStreamResource(logger, testStore)
	require.NoError(t, err)

	t.Run("Successfully read empty log stream", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-empty"

		req := &mcp.ReadResourceRequest{
			Params: mcp.ReadResourceRequestParams{
				URI: "mcp://tests/" + testID + "/logs",
			},
		}

		result, err := resource.ReadTestLogStream(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Contents, 1)
		require.Equal(t, "mcp://tests/"+testID+"/logs", result.Contents[0].URI)
		require.Equal(t, "application/json", result.Contents[0].MIMEType)

		// Verify JSON content
		var payload map[string]any
		err = json.Unmarshal([]byte(result.Contents[0].Text), &payload)
		require.NoError(t, err)
		require.NotEmpty(t, payload)
		require.False(t, payload["meta"].(map[string]any)["final"].(bool))
	})

	t.Run("Successfully read log stream with events", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-with-events"

		// Add events to the store
		testStore.AddEvent(testID, entity.LogEvent{
			Event: "progress",
			Data:  `{"status":"started"}`,
		})
		testStore.AddEvent(testID, entity.LogEvent{
			Event: "log",
			Data:  `{"message":"test running"}`,
		})

		req := &mcp.ReadResourceRequest{
			Params: mcp.ReadResourceRequestParams{
				URI: "mcp://tests/" + testID + "/logs",
			},
		}

		result, err := resource.ReadTestLogStream(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Contents, 1)

		// Verify JSON content
		var payload map[string]any
		err = json.Unmarshal([]byte(result.Contents[0].Text), &payload)
		require.NoError(t, err)

		events := payload["events"].([]any)
		require.Len(t, events, 2)

		firstEvent := events[0].(map[string]any)
		require.Equal(t, "progress", firstEvent["Event"].(string))

		secondEvent := events[1].(map[string]any)
		require.Equal(t, "log", secondEvent["Event"].(string))
	})

	t.Run("Successfully read completed log stream", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-completed"

		// Add event and complete the stream
		testStore.AddEvent(testID, entity.LogEvent{
			Event: "progress",
			Data:  `{"status":"completed"}`,
		})
		testStore.CompleteStream(testID)

		req := &mcp.ReadResourceRequest{
			Params: mcp.ReadResourceRequestParams{
				URI: "mcp://tests/" + testID + "/logs",
			},
		}

		result, err := resource.ReadTestLogStream(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify JSON content
		var payload map[string]any
		err = json.Unmarshal([]byte(result.Contents[0].Text), &payload)
		require.NoError(t, err)

		require.True(t, payload["meta"].(map[string]any)["final"].(bool))
	})

	t.Run("Error with invalid URI format", func(t *testing.T) {
		ctx := context.Background()

		req := &mcp.ReadResourceRequest{
			Params: mcp.ReadResourceRequestParams{
				URI: "invalid://format",
			},
		}

		result, err := resource.ReadTestLogStream(ctx, req)

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("Error with non-existent stream", func(t *testing.T) {
		ctx := context.Background()

		req := &mcp.ReadResourceRequest{
			Params: mcp.ReadResourceRequestParams{
				URI: "mcp://tests/non-existent/logs",
			},
		}

		result, err := resource.ReadTestLogStream(ctx, req)

		require.Error(t, err)
		require.Nil(t, result)
	})
}
