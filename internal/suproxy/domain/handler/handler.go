package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type SuproxyController struct {
	logger *slog.Logger
}

func NewSuproxyController(logger *slog.Logger) (*SuproxyController, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger: logger,
	}, nil
}

func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	var request entity.Request

	if err := c.BindJSON(request); err != nil {
		return
	}

	c.JSON(http.StatusOK, 200)
}
