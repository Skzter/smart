package main

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

func main() {
	logger := slog.Default()
	if logger == nil {
		panic("logger is nil")
	}

	handler, err := handler.NewSuproxyController(logger)
	if err != nil {
		panic(err)
	}

	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", handler.PostOfferlist)
	}

	if err := router.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
