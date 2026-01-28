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

	tests := []struct {
		name  string
		input entity.ReadTestLogStreamIn
		setup func(store.TestLogStreamStore)
		check func(*testing.T, *entity.ReadTestLogStreamResponse)
	}{
		{
			name: "Successfully read empty log stream",
			input: entity.ReadTestLogStreamIn{
				TestId: "test-123",
			},
			check: func(t *testing.T, output *entity.ReadTestLogStreamResponse) {
				require.NotNil(t, output)
				require.Empty(t, output.Content)
				require.False(t, output.Meta["exists"].(bool))
			},
		},
		{
			name: "Successfully read log stream with events",
			input: entity.ReadTestLogStreamIn{
				TestId: "test-456",
			},
			setup: func(s store.TestLogStreamStore) {
				s.AddEvent("test-456", entity.LogEvent{
					Event: "progress",
					Data:  `{"status":"started"}`,
				})
				s.AddEvent("test-456", entity.LogEvent{
					Event: "log",
					Data:  `{"message":"test running"}`,
				})
			},
			check: func(t *testing.T, output *entity.ReadTestLogStreamResponse) {
				require.NotNil(t, output)
				require.Len(t, output.Content, 2)
				require.Equal(t, "progress", output.Content[0].Event)
				require.Equal(t, "log", output.Content[1].Event)
				require.False(t, output.Meta["final"].(bool))
			},
		},
		{
			name: "Successfully read completed log stream",
			input: entity.ReadTestLogStreamIn{
				TestId: "test-789",
			},
			setup: func(s store.TestLogStreamStore) {
				s.AddEvent("test-789", entity.LogEvent{
					Event: "progress",
					Data:  `{"status":"completed"}`,
				})
				s.CompleteStream("test-789")
			},
			check: func(t *testing.T, output *entity.ReadTestLogStreamResponse) {
				require.NotNil(t, output)
				require.Len(t, output.Content, 1)
				require.True(t, output.Meta["final"].(bool))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testStore := store.NewTestLogStreamStore()
			defer testStore.Shutdown()

			tool, err := NewReadTestLogsTool(logger, testStore)
			require.NoError(t, err)

			if test.setup != nil {
				test.setup(testStore)
			}

			ctx := context.Background()
			_, output, err := tool.ReadTestLogStream(ctx, nil, test.input)

			require.NoError(t, err)
			test.check(t, output)
		})
	}
}
