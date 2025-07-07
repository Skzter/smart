package application

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/web"
)

// Initializes the HTTP server, sets up API routes and serves static files.
// Registers endpointa and serves frontend assets from the embedded dist directory.
func SetupRoutes(cfg *config.Config) {
	router := gin.Default()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	controller, err := handler.NewAutotesterController(logger, cfg)

	if err != nil {
		logger.Error(err.Error())
		return
	}

	router.POST("/api/v1/chat", func(c *gin.Context) {
		controller.HandleChatRequest(c)
	})

	staticFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		logger.Error(err.Error())
		return
	}

	router.StaticFS("/", http.FS(staticFS))

	err = router.Run(cfg.Port)
	if err != nil {
		logger.Error(err.Error())
		return
	}
}
