package application

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
)

// McpServer wraps the MCP SDK server and provides dependency injection
type McpServer struct {
	server *mcp.Server
	config *config.Config
}

// NewMcpServer creates a new MCP server instance with all dependencies
func NewMcpServer(server *mcp.Server, cfg *config.Config) (*McpServer, error) {
	wrapper := &McpServer{
		server: server,
		config: cfg,
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
	// TODO: Tools registrieren
	// - GetTemplate
	// - GenerateTest
	// - RunTest
	// - GetTestStatus (später)

	return nil
}

// Server returns the underlying mcp.Server (for testing)
func (m *McpServer) Server() *mcp.Server {
	return m.server
}
