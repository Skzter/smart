//go:build integration

package tools

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
)

func TestReadTestLogsTool_Integration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a test store
	testStore := store.NewTestLogStreamStore()
	defer testStore.Shutdown()

	// Create the tool
	tool, err := NewReadTestLogsTool(logger, testStore)
	require.NoError(t, err)

	t.Run("Successfully read empty log stream", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-123"

		_, output, err := tool.ReadTestLogStream(ctx, nil, entity.ReadTestLogStreamIn{
			TestId: testID,
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Empty(t, output.Content)
		require.False(t, output.Meta["exists"].(bool))
	})

	t.Run("Successfully read log stream with events", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-456"

		// Add events to the store
		testStore.AddEvent(testID, entity.LogEvent{
			Event: "progress",
			Data:  `{"status":"started"}`,
		})
		testStore.AddEvent(testID, entity.LogEvent{
			Event: "log",
			Data:  `{"message":"test running"}`,
		})

		_, output, err := tool.ReadTestLogStream(ctx, nil, entity.ReadTestLogStreamIn{
			TestId: testID,
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Len(t, output.Content, 2)
		require.Equal(t, "progress", output.Content[0].Event)
		require.Equal(t, "log", output.Content[1].Event)
		require.True(t, output.Meta["exists"].(bool))
	})

	t.Run("Successfully read completed log stream", func(t *testing.T) {
		ctx := context.Background()
		testID := "test-789"

		testStore.AddEvent(testID, entity.LogEvent{
			Event: "progress",
			Data:  `{"status":"completed"}`,
		})

		testStore.CompleteStream(testID)

		_, output, err := tool.ReadTestLogStream(ctx, nil, entity.ReadTestLogStreamIn{
			TestId: testID,
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Len(t, output.Content, 1)
		require.True(t, output.Meta["final"].(bool))
	})
}
