// Package resource provides MCP resource implementations for the autotester system.
package resource

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestLogStreamResource handles MCP resource requests for test execution logs.
// It serves as the data provider for test logs that LLMs can access through a dedicated
// read_test_logs tool. The resource supports signaling whether the test execution is still
// running (final: false) or completed (final: true), allowing LLMs to make intelligent
// decisions about polling for new logs.
type TestLogStreamResource struct {
	logger               *slog.Logger
	autotesterAPIService service.AutotesterAPIService
}

// NewTestLogStreamResource creates a new TestLogStreamResource instance.
// It validates that all required dependencies (logger and autotesterAPIService) are provided.
//
// Returns an error if any dependency is nil.
func NewTestLogStreamResource(logger *slog.Logger, autotesterAPIService service.AutotesterAPIService) (*TestLogStreamResource, error) {
	if err := assert.NotNil(logger, autotesterAPIService); err != nil {
		return nil, err
	}
	return &TestLogStreamResource{
		logger:               logger,
		autotesterAPIService: autotesterAPIService,
	}, nil
}

// ReadTestLogStream reads test execution logs for a given testId from the MCP resource request.
// It extracts the testId from the resource URI (format: mcp://tests/{testId}/logs) and fetches
// the accumulated logs from the autotester service.
//
// The response includes a "final" metadata flag:
//   - final: false indicates the test is still running and more logs may be available
//   - final: true indicates the test has completed and all logs are included
//
// This allows LLMs to make intelligent decisions about whether to continue polling
// for new logs or if the test execution is fully complete.
func (ls *TestLogStreamResource) ReadTestLogStream(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI
	// TODO: service anschließen
	return &mcp.ReadResourceResult{
		Meta: mcp.Meta{"final": true}, // dynamisch setzen
		Contents: []*mcp.ResourceContents{
			{
				URI:  uri,
				Text: "Hey",
			},
		},
	}, nil
}
