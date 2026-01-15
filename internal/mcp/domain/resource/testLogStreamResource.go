// Package resource provides MCP resource implementations for the autotester system.
package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestLogStreamResource handles MCP resource requests for test execution logs.
// It serves as the data provider for test logs that LLMs can access through a dedicated
// read_test_logs tool. The resource supports signaling whether the test execution is still
// running (final: false) or completed (final: true), allowing LLMs to make intelligent
// decisions about polling for new logs.
type TestLogStreamResource struct {
	logger *slog.Logger
	store  store.TestLogStreamStore
}

// NewTestLogStreamResource creates a new TestLogStreamResource instance.
// It validates that all required dependencies (logger and autotesterAPIService) are provided.
//
// Returns an error if any dependency is nil.
func NewTestLogStreamResource(logger *slog.Logger, store store.TestLogStreamStore) (*TestLogStreamResource, error) {
	if err := assert.NotNil(logger, store); err != nil {
		return nil, err
	}
	return &TestLogStreamResource{
		logger: logger,
		store:  store,
	}, nil
}

// ReadTestLogStream reads test execution logs for a given testId from the MCP resource request.
// It extracts the testId from the resource URI (format: mcp://tests/{testId}/logs) and fetches
// the accumulated logs from the autotester service.
func (ls *TestLogStreamResource) ReadTestLogStream(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	// expected format: mcp://tests/{testId}/logs
	const prefix = "mcp://tests/"
	const suffix = "/logs"
	if !strings.HasPrefix(uri, prefix) || !strings.HasSuffix(uri, suffix) {
		ls.logger.Warn("invalid test logs uri", "uri", uri)
		return nil, nil
	}

	testId := strings.TrimSuffix(strings.TrimPrefix(uri, prefix), suffix)
	ls.logger.Debug("extracted test id from uri", "testId", testId)

	stream, exists := ls.store.GetStream(testId)
	if !exists {
		return nil, fmt.Errorf("stream for testId: %s doesnt exist", testId)
	}

	events := stream.GetEvents()

	payload := map[string]any{
		"events": events,
		"meta": map[string]any{
			"final": stream.IsCompleted(),
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		ls.logger.Error("failed to marshal log events", "err", err)
		return nil, err
	}

	contents := &mcp.ResourceContents{
		URI:      uri,
		MIMEType: "application/json",
		Text:     string(b),
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{contents},
	}, nil
}
