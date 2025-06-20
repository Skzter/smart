package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/web"
)

func main() {
	router := gin.Default()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	controller, err := handler.NewAutotesterController(logger)

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

	// Compatibility with Nginx Proxy
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// If the request is for /api/assets/, serve from assets directory
		if strings.HasPrefix(path, "/api/assets/") {
			// Remove /api prefix to get the actual file path
			filePath := strings.TrimPrefix(path, "/api")
			c.Request.URL.Path = filePath
		}

		// Serve the file using the static filesystem
		fileServer := http.FileServer(http.FS(staticFS))
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	err = router.Run(":8081")
	if err != nil {
		logger.Error(err.Error())
		return
	}
}
