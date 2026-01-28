package application

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/resource"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/tools"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// McpServer wraps the MCP SDK server and provides dependency injection
type McpServer struct {
	server            *mcp.Server
	autotesterService service.AutotesterAPIService
	store             store.TestLogStreamStore
	logger            *slog.Logger
}

// NewMcpServer creates a new MCP server instance with all dependencies
func NewMcpServer(
	server *mcp.Server,
	autotesterService service.AutotesterAPIService,
	store store.TestLogStreamStore,
	logger *slog.Logger,
) (*McpServer, error) {
	if err := assert.NotNil(server, autotesterService, store, logger); err != nil {
		return nil, err
	}

	wrapper := &McpServer{
		server:            server,
		autotesterService: autotesterService,
		store:             store,
		logger:            logger,
	}

	if err := wrapper.registerTools(); err != nil {
		return nil, err
	}

	if err := wrapper.registerResources(); err != nil {
		return nil, err
	}

	return wrapper, nil
}

// Run starts the MCP server with the given transport
func (m *McpServer) Run(ctx context.Context, transport mcp.Transport) error {
	return m.server.Run(ctx, transport)
}

// ShutdownComponents stops background components started by the MCP wrapper.
//
// Note: This currently only stops the stores background routine
func (m *McpServer) ShutdownComponents() {
	m.store.Shutdown()
}

// registerTools registers all available tools with the MCP server
func (m *McpServer) registerTools() error {
	getTemplateTool, err := tools.NewGetTemplateTool(m.logger, m.autotesterService)
	if err != nil {
		return err
	}

	runTestTool, err := tools.NewRunTestTool(m.logger, m.autotesterService)
	if err != nil {
		return err
	}

	generateTestTool, err := tools.NewGenerateTestTool(m.logger, m.autotesterService)
	if err != nil {
		return err
	}
	readTestLogsTool, err := tools.NewReadTestLogsTool(m.logger, m.store)
	if err != nil {
		return err
	}

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_template",
		Description: "Retrieves the test generation template from the autotester backend",
	}, getTemplateTool.GetTemplate)

	m.logger.Debug("Registered tool: get_template")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "generate_test",
		Description: "takes prompt, sends it to backend and receives feedback to display in frontend",
	}, generateTestTool.GenerateTest)

	m.logger.Debug("Registered tool: generate_test")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "run_test",
		Description: "Runs test based on testId, returns either success or failed",
	}, runTestTool.RunTest)

	m.logger.Debug("Registered tool: run_test")

	mcp.AddTool(m.server, &mcp.Tool{
		Name: "read_testLogStream",
		Description: `Reads the test log stream for a given testId and returns JSON with "events" and "meta.final".
	Avoid excessive polling — prefer waiting 2–5 seconds between attempts. If using a backoff strategy, ensure a reasonable maximum (e.g. 30s) to avoid waiting too long.
	Inspect "meta.final" to determine whether the stream is complete; poll again until "meta.final" is true.
	Note: after invoking "run_test" it may take some time for the stream to be processed and become available.`,
	}, readTestLogsTool.ReadTestLogStream)

	m.logger.Debug("Registered tool: read_testLogStream")

	return nil
}

// registerResources registers all available resources with the MCP server
func (m *McpServer) registerResources() error {
	testLogStreamResource, err := resource.NewTestLogStreamResource(m.logger, m.store)
	if err != nil {
		return err
	}

	m.server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "testlog_stream",
		Description: "Provides access to test execution logs for a specific test. The URI must include the `{testId}` path parameter; returns JSON (events and metadata, e.g. final flag).",
		URITemplate: "mcp://tests/{testId}/logs",
		MIMEType:    "application/json",
	}, testLogStreamResource.ReadTestLogStream)

	m.logger.Debug("Registered resource template: mcp://tests/123/logs")
	return nil
}

// Server returns the underlying mcp.Server (for testing)
func (m *McpServer) Server() *mcp.Server {
	return m.server
}
