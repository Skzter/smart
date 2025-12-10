package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
)

func main() {
	cfg, err := config.LoadMCPConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mcpServer, err := InitializeMcpServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MCP server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down MCP server...")
		cancel()
	}()

	log.Println("Starting MCP server on stdio...")

	transport := mcp.StdioTransport{}

	if err := mcpServer.Run(ctx, &transport); err != nil {
		panic(err)
	}
}
