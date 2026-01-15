package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ReadTestLogStreamTool is a tool for reading test execution logs from a specific test.
type ReadTestLogStreamTool struct {
	logger *slog.Logger
	store  store.TestLogStreamStore
}

// NewReadTestLogStreamTool creates a new ReadTestLogStreamTool
func NewReadTestLogStreamTool(logger *slog.Logger, store store.TestLogStreamStore) (*ReadTestLogStreamTool, error) {
	if err := assert.NotNil(logger, store); err != nil {
		return nil, err
	}
	return &ReadTestLogStreamTool{
		logger: logger,
		store:  store,
	}, nil
}

// ReadTestLogStream reads test execution logs for a given test ID.
// It returns the accumulated logs along with metadata about the stream state.
func (rt *ReadTestLogStreamTool) ReadTestLogStream(
	ctx context.Context,
	request *mcp.CallToolRequest,
	input entity.ReadTestLogStreamIn,
) (
	result *mcp.CallToolResult,
	output *entity.ReadTestLogStreamResponse,
	_ error,
) {
	rt.logger.Debug("ReadTestLogStream tool called", "test_id", input.TestId)
	stream, exists := rt.store.GetStream(input.TestId)
	if !exists {
		return nil, &entity.ReadTestLogStreamResponse{
			Content: []entity.LogEvent{},
			Meta: map[string]any{
				"exists": false,
			},
		}, nil
	}

	events := stream.GetEvents()
	return nil, &entity.ReadTestLogStreamResponse{
		Content: events,
		Meta: map[string]any{
			"final": stream.IsCompleted(),
		},
	}, nil
}
