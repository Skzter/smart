package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
)

func main() {
	cfg, err := config.LoadMCPConfig()
	if err != nil {
		panic(err)
	}

	mcpServer, err := InitializeMcpServer(cfg)
	if err != nil {
		panic(err)
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

	log.Println("Starting MCP server on HTTP :8084...")

	// Create SSE-Handler erzeugen (Transport Layer)
	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer.Server()
	}, nil)

	// HTTP-Server configuration
	srv := &http.Server{
		Addr:              ":8084",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Server shuts down
	go func() {
		<-ctx.Done()
		log.Println("Shutting down MCP HTTP server...")
		_ = srv.Shutdown(context.Background())
	}()

	// Start HTTP-Server
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
