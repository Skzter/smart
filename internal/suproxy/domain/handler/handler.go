package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// SuproxyController handles the HTTP requests for the Suproxy service
type SuproxyController struct {
	logger *slog.Logger
	config *config.Config
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(logger *slog.Logger, config *config.Config) (*SuproxyController, error) {
	if err := assert.NotNil(logger, config); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger: logger,
		config: config,
	}, nil
}

// PostOfferlist handles the POST request to the /api/v1/Offerlist endpoint
func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	var request entity.Request

	if err := c.BindJSON(&request); err != nil {
		s.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	c.JSON(http.StatusOK, 200)
}
