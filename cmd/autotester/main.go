package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
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

	// staticFS, _ := fs.Sub(website.DistFS, "dist")
	// router.StaticFS("/", http.FS(staticFS))

	err = router.Run(":8081")

	if err != nil {
		return
	}
}
