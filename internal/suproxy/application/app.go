package application

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Run initializes the application, sets up the logger, handler, and router,
func Run() {
	logger := slog.Default()
	if logger == nil {
		panic("logger is nil")
	}

	cfg, err := config.LoadAppConfig()
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	handler, err := handler.NewSuproxyController(logger, cfg)
	if err != nil {
		logger.Error("failed to create handler", "error", err)
		return
	}

	router := setupRouter(handler)

	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("failed to run server", "error", err)
		return
	}
}

// setupRouter initializes the Gin router and sets up the routes for the API
func setupRouter(h *handler.SuproxyController) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	return router
}
