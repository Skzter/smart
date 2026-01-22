package application

import (
	"net/http"
	"os"

	ddgin "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

// NewRouter initializes the Gin router and sets up the routes for the API
func NewRouter(h *handler.SuproxyController) *gin.Engine {
	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(ddgin.Middleware(os.Getenv("DD_SERVICE")))
	router.Use(corsMiddleware())

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	debugGroup := router.Group("/debug", pprofAuthMiddleware())
	pprof.RouteRegister(debugGroup, "pprof")

	return router
}

func pprofAuthMiddleware() gin.HandlerFunc {
	password := os.Getenv("PPROF_AUTH_PASSWORD")
	if password == "" {
		password = "smart-qa"
	}

	return gin.BasicAuth(gin.Accounts{
		"admin": password,
	})
}

// corsMiddleware handles CORS headers
func corsMiddleware() gin.HandlerFunc {
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
