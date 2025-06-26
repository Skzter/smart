package application

import (
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// App represents the application structure
type App struct {
	logger *slog.Logger
	router *gin.Engine
}

// newApp creates a new instance of App
func newApp() (*App, error) {
	logger := slog.Default()
	if logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	handler, err := handler.NewSuproxyController(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	router := setupRouter(handler)
	if router == nil {
		return nil, fmt.Errorf("router is nil")
	}
	return &App{
		logger: logger,
		router: router,
	}, nil
}

// Run starts the application and listens on port 127.0.0.1:8080
func Run() {
	app, err := newApp()
	if err != nil {
		panic(err)
	}

	if err := app.router.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}

// setupRouter initializes the Gin router and sets up the routes for the API
func setupRouter(h *handler.SuproxyController) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	return router
}
