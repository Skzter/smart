package application

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/web"
)

// Initializes the HTTP server, sets up API routes and serves static files.
// Registers endpointa and serves frontend assets from the embedded dist directory.
func NewRouter(logger *slog.Logger, controller *handler.AutotesterController) *gin.Engine {
	router := gin.Default()

	router.POST("/api/v1/chat", controller.HandleChatRequest)

	staticFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	router.StaticFS("/", http.FS(staticFS))

	return router
}
