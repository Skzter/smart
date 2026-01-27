package application

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"

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

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_template",
		Description: "Fetches the prompt template to structure test generation requests so that LLM can use it",
	}, getTemplateTool.GetTemplate)

	m.logger.Info("Registered tool: get_template")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "generate_test",
		Description: "Takes prompt from user, sends it to backend to validate and then generates test code, if validation fails, it returns validation message",
	}, generateTestTool.GenerateTest)

	m.logger.Info("Registered tool: generate_test")

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "run_test",
		Description: "At first it saves test code provided by the user and runs it afterwards, then it returns message to the user whether test started succeeded or failed to start",
	}, runTestTool.RunTest)

	m.logger.Info("Registered tool: run_test")

	return nil
}

// Server returns the underlying mcp.Server (for testing)
func (m *McpServer) Server() *mcp.Server {
	return m.server
}
