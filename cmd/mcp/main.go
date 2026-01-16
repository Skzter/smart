package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	ddgin "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/tracing"
)

func main() {
	cfg, err := config.LoadMCPConfig()
	if err != nil {
		panic(err)
	}

	tracer, shutdown, err := tracing.Setup("mcp")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := shutdown(); err != nil {
			panic(err)
		}
	}()

	mcpServer, err := InitializeMcpServer(cfg, tracer)
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

	log.Printf("Starting MCP server on HTTP %s\n", cfg.Port)

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer.Server()
	}, nil)

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(ddgin.Middleware(os.Getenv("DD_SERVICE")))

	mcpHandler := func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}

	router.Group("/mcp", mcpHandler)

	go func() {
		<-ctx.Done()
		log.Println("Shutting down MCP HTTP server...")
		os.Exit(0)
	}()

	if err := router.Run(cfg.Port); err != nil {
		panic(err)
	}
}
