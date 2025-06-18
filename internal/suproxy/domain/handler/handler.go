package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

type SuproxyController struct {
	PostOfferlist func(c *gin.Context)
}

func PostOfferlist(c *gin.Context) {
	var request entity.Request

	if err := c.BindJSON(request); err != nil {
		return
	}
	c.JSON(http.StatusOK, 200)
}
