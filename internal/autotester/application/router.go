package application

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/web"
)

// NewRouter initializes the HTTP server, sets up API routes and serves static files.
// Registers endpointa and serves frontend assets from the embedded dist directory.
func NewRouter(logger *slog.Logger, controller *handler.AutotesterController) (*gin.Engine, error) {
	router := gin.Default()

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/chat", controller.HandleChatRequest)
		apiV1.GET("/chats/:UserID", controller.HandleGetUserChats)
		apiV1.GET("/template", controller.HandleGetTemplate)
		apiV1.POST("/saveLocal", controller.HandleSaveLocalRequest)
		apiV1.DELETE("/deleteLocal", controller.HandleDeleteLocalRequest)
		apiV1.POST("/run", controller.HandleRunContainer)
	}

	router.GET("/auth_config.json", func(c *gin.Context) {
		c.FileFromFS("/auth_config.json", http.FS(web.Auth0Config))
	})

	assetsFS, err := fs.Sub(web.DistFS, "dist/assets")
	if err != nil {
		return nil, err
	}
	router.StaticFS("/assets", http.FS(assetsFS))

	router.GET("/", func(c *gin.Context) {
		indexHTML, err := web.DistFS.ReadFile("dist/index.html")
		if err != nil {
			logger.Error("Failed to read embedded index.html", "error", err)
			c.String(http.StatusInternalServerError, "Internal Server Error")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	logger.Info("Router initialized")
	return router, nil
}
