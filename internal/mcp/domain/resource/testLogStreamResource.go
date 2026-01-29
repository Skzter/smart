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

// TestLogStreamResource implements an MCP resource that provides test execution logs to
// MCP clients. Clients receive a JSON payload containing the accumulated SSE events and
// a `meta.final` boolean indicating whether the test run has completed.
type TestLogStreamResource struct {
	logger *slog.Logger
	store  store.TestLogStreamStore
}

// NewTestLogStreamResource creates a new TestLogStreamResource instance.
// It validates that required dependencies (logger and store) are provided and returns an
// error if any dependency is nil.
func NewTestLogStreamResource(logger *slog.Logger, store store.TestLogStreamStore) (*TestLogStreamResource, error) {
	if err := assert.NotNil(logger, store); err != nil {
		return nil, err
	}
	return &TestLogStreamResource{
		logger: logger,
		store:  store,
	}, nil
}

// ReadTestLogStream handles MCP read requests for a test's log stream. The request URI must
// use the format `mcp://tests/{testId}/logs`. The method extracts the `testId`, fetches the
// accumulated events from the store and returns JSON with `events` and `meta.final`.
func (ls *TestLogStreamResource) ReadTestLogStream(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	uri := req.Params.URI

	// expected format: mcp://tests/{testId}/logs
	const prefix = "mcp://tests/"
	const suffix = "/logs"
	if !strings.HasPrefix(uri, prefix) || !strings.HasSuffix(uri, suffix) {
		ls.logger.Warn("invalid test logs uri", "uri", uri)
		return nil, fmt.Errorf("invalid test logs uri: %s", uri)
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

	serializedPayload, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		ls.logger.Error("failed to marshal log events", "err", err)
		return nil, err
	}

	contents := &mcp.ResourceContents{
		URI:      uri,
		MIMEType: "application/json",
		Text:     string(serializedPayload),
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{contents},
	}, nil
}
