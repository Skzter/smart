package application

import (
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
func NewApp() (*App, error) {
	logger := slog.Default()
	if logger == nil {
		panic("logger is nil")
	}

	handler, err := handler.NewSuproxyController(logger)
	if err != nil {
		panic(err)
	}

	router := SetupRouter(handler)
	if router == nil {
		panic("router is nil")
	}
	return &App{
		logger: logger,
		router: router,
	}, nil
}

func Run() {
	app, err := NewApp()
	if err != nil {
		panic(err)
	}

	if err := app.router.Run(":8080"); err != nil {
		panic(err)
	}
}

// setupRouter initializes the Gin router and sets up the routes for the API
func SetupRouter(h *handler.SuproxyController) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	return router
}
