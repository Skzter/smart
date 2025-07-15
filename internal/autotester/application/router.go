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
func NewRouter(logger *slog.Logger, controller *handler.AutotesterController) *gin.Engine {
	router := gin.Default()

	apiV1 := router.Group("/v1")
	{
		apiV1.POST("/chat", controller.HandleChatRequest)
	}

	router.GET("/auth_config.json", func(c *gin.Context) {
		c.FileFromFS("/auth_config.json", http.FS(web.Auth0Config))
	})

	assetsFS, err := fs.Sub(web.DistFS, "dist/assets")
	if err != nil {
		logger.Error(err.Error())
		return nil
	}
	router.StaticFS("/assets", http.FS(assetsFS))

	router.LoadHTMLFiles("web/dist/index.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	logger.Info("Router initialized")
	return router
}
