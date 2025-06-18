package main

import (
	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
)

func main() {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", handler.PostOfferlist)
	}

	if err := router.Run("127.0.0.1:8080"); err != nil {
		panic(err)
	}
}
