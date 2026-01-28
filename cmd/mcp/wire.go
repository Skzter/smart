//go:build wireinject

package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/wire"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/application"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/store"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/logger"
)

// InitializeMcpServer initializes the mcp server with all dependencies
func InitializeMcpServer(cfg *config.Mcp) (*application.McpServer, error) {
	wire.Build(
		ProvideHTTPClient,
		ProvideLogger,
		ProvideBaseURL,

		repository.NewAutotesterAPIRepository,

		service.NewAutotesterAPIService,
		service.NewJWTAuthentification,

		store.NewTestLogStreamStore,

		ProvideMcpServer,

		application.NewMcpServer,
	)
	return nil, nil
}

// ProvideLogger provides a new logger.
func ProvideLogger(cfg *config.Mcp) *slog.Logger {
	return logger.NewLogger(cfg.LogLevel, cfg.LogFilePath)
}

// ProvideHTTPClient creates a configured HTTP client from config
func ProvideHTTPClient(cfg *config.Mcp) *http.Client {
	return &http.Client{
		Timeout: time.Duration(cfg.HttpClient.TimeoutSeconds) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.HttpClient.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.HttpClient.MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Duration(cfg.HttpClient.IdleConnTimeoutSeconds) * time.Second,
		},
	}
}

// ProvideBaseURL extracts the base URL from config
func ProvideBaseURL(cfg *config.Mcp) string {
	return cfg.AutotesterAPIBaseURL
}

// ProvideMcpServer creates the underlying mcp.Server from config
func ProvideMcpServer(cfg *config.Mcp) *mcp.Server {
	serverOpts := mcp.ServerOptions{HasResources: true}
	impl := &mcp.Implementation{
		Name:    cfg.McpServerImplementation.Name,
		Version: cfg.McpServerImplementation.Version,
	}

	return mcp.NewServer(impl, &serverOpts)
}
