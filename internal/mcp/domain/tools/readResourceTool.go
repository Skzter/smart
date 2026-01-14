package tools

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ReadTestLogStreamTool is a tool for reading test execution logs from a specific test.
type ReadTestLogStreamTool struct {
	logger *slog.Logger
}

// NewReadTestLogStreamTool creates a new ReadTestLogStreamTool
func NewReadTestLogStreamTool(logger *slog.Logger) (*ReadTestLogStreamTool, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}
	return &ReadTestLogStreamTool{
		logger: logger,
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
	rt.logger.Debug("ReadTestLogStream tool called", "test_id", input.TestID, "cursor", input.Cursor)

	return nil, &entity.ReadTestLogStreamResponse{
		Content: "Test logs for test ID: " + input.TestID,
		Meta: map[string]any{
			"cursor": input.Cursor,
			"final":  true,
		},
	}, nil
}
