package application

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/resource"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/tools"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// McpServer wraps the MCP SDK server and provides dependency injection
type McpServer struct {
	server            *mcp.Server
	autotesterService service.AutotesterAPIService
	logger            *slog.Logger
}

// NewMcpServer creates a new MCP server instance with all dependencies
func NewMcpServer(server *mcp.Server, autotesterService service.AutotesterAPIService, logger *slog.Logger) (*McpServer, error) {
	if err := assert.NotNil(server, autotesterService, logger); err != nil {
		return nil, err
	}

	wrapper := &McpServer{
		server:            server,
		autotesterService: autotesterService,
		logger:            logger,
	}

	if err := wrapper.registerTools(); err != nil {
		return nil, err
	}

	return wrapper, nil
}

// Run starts the MCP server with the given transport
func (m *McpServer) Run(ctx context.Context, transport mcp.Transport) error {
	return m.server.Run(ctx, transport)
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

	testLogStreamResource, err := resource.NewTestLogStreamResource(m.logger, m.autotesterService)
	if err != nil {
		return err
	}

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_template",
		Description: "Retrieves the test generation template from the autotester backend",
	}, getTemplateTool.GetTemplate)

	m.logger.Info("Registered tool: get_template")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "generate_test",
		Description: "takes prompt, sends it to backend and receives feedback to display in frontend",
	}, generateTestTool.GenerateTest)

	m.logger.Info("Registered tool: generate_test")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "run_test",
		Description: "Runs test based on testId, returns either success or failed",
	}, runTestTool.RunTest)

	m.logger.Info("Registered tool: run_test")

	m.server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "testlog_stream",
		Description: "Provides access to test execution logs",
		URITemplate: "mcp://tests/{testId}/logs",
		MIMEType:    "text/plain",
	}, testLogStreamResource.ReadTestLogStream)

	m.logger.Info("Registered resource template: mcp://tests/123/logs")
	return nil
}

// Server returns the underlying mcp.Server (for testing)
func (m *McpServer) Server() *mcp.Server {
	return m.server
}
