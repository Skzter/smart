package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// SuproxyController handles the HTTP requests for the Suproxy service
type SuproxyController struct {
	logger    *slog.Logger
	validator *service.Validator
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(logger *slog.Logger, validator *service.Validator) (*SuproxyController, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:    logger,
		validator: validator,
	}, nil
}

// PostOfferlist handles the POST request to the /api/v1/Offerlist endpoint
func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	var request entity.Request

	if err := c.BindJSON(&request); err != nil {
		s.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	err := s.validator.Validate(request.Request)
	if err != nil {
		s.logger.Error("Validation failed", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, 200)
}
