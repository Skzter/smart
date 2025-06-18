package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
)

func main() {
	router := gin.Default()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	router.POST("/api/v1/chat", func(c *gin.Context) {
		handler.HandleChatRequest(c, logger, ctx)
	})

	err := router.Run(":8080")

	if err != nil {
		return
	}
}
