package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	ddgin "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/gin-gonic/gin"
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

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer.Server()
	}, nil)

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(ddgin.Middleware(os.Getenv("DD_SERVICE")))
	router.Use(mcpServer.JwtExtraction.JWTExtractionIntoContext())

	router.POST("/mcp", func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	})

	httpServer := &http.Server{
		Addr:    cfg.Port,
		Handler: router,

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("Starting MCP server", "port", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received")
	mcpServer.ShutdownComponents()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Shutting down HTTP server")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", "error", err)
	}

	slog.Info("MCP server exited gracefully")
}
