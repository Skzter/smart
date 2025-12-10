//go:build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
)

// InitializeMcpServer initializes the mcp server with all dependencies
func InitializeMcpServer(cfg *config.Config) (*application.McpServer, error) {
	wire.Build(
		ProvideMcpServer,
		application.NewMcpServer,
	)
	return nil, nil
}

// ProvideMcpServer creates the underlying mcp.Server from config
func ProvideMcpServer(cfg *config.Config) *mcp.Server {
	impl := &mcp.Implementation{
		Name:    cfg.McpServerImplementation.Name,
		Version: cfg.McpServerImplementation.Version,
	}

	return mcp.NewServer(impl, nil)
}
