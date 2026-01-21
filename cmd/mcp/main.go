package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	tracer, shutdownTracer, err := tracing.Setup("mcp")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := shutdownTracer(); err != nil {
			log.Println("tracer shutdown error:", err)
		}
	}()

	mcpServer, err := InitializeMcpServer(cfg, tracer)
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return mcpServer.Server()
	}, nil)

	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(ddgin.Middleware(os.Getenv("DD_SERVICE")))

	router.Any("/mcp/*any", func(c *gin.Context) {
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
		log.Printf("Starting MCP server on %s\n", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("MCP server exited gracefully")
}
