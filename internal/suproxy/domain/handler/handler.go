package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	validator "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// SuproxyController handles the HTTP requests for the Suproxy service
type SuproxyController struct {
	logger *slog.Logger
	config *config.Config
	client *http.Client
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(logger *slog.Logger, config *config.Config) (*SuproxyController, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger: logger,
		config: config,
		client: &http.Client{},
	}, nil
}

// PostOfferlist handles the POST request to the /api/v1/Offerlist endpoint
func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	request := c.Request

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8083/api/v1/offers", request.Body)
	if err != nil {
		s.logger.Error("Failed to create request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "1Invalid request body"})
		return
	}

	defer func() {
		if err := req.Body.Close(); err != nil {
			s.logger.Error("Failed to close response body", "error", err)
		}
	}()

	req.Header = request.Header

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "2Invalid request body"})
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Failed to close response body", "error", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "4Invalid request body"})
		return
	}

	// !TODO: REMOVE THIS - THIS IS FOR TESTING ONLY

	supresp := entity.SupplierOfferResponse{
		HTTPStatusCode: resp.StatusCode,
		Data:           body,
	}
	go s.handleRequest(c, supresp)

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (s *SuproxyController) handleRequest(ctx context.Context, data entity.SupplierOfferResponse) {
	val := validator.NewValidator(s.logger, s.config)
	if err := val.Validate(ctx, &data); err != nil {
		s.logger.Error(err.Error())
	}
}
