package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	logger    *slog.Logger
	config    *config.Config
	client    *http.Client
	validator validator.Validator
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(logger *slog.Logger, config *config.Config) (*SuproxyController, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:    logger,
		config:    config,
		client:    &http.Client{},
		validator: *validator.NewValidator(logger, config),
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

	req, err := http.NewRequest(http.MethodPost, request.Destination, bytes.NewBuffer([]byte(request.Request)))
	if err != nil {
		s.logger.Error("Failed to create request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	defer func() {
		if err := req.Body.Close(); err != nil {
			s.logger.Error("Failed to close response body", "error", err)
		}
	}()

	for key, value := range request.Header {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Failed to send request", "error", resp.StatusCode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Internal"})
		return
	}

	go s.handleRequest(c, &body)

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (s *SuproxyController) handleRequest(ctx context.Context, data *[]byte) {
	var mappedData map[string]any
	if err := json.Unmarshal(*data, &mappedData); err != nil {
		s.logger.Error(err.Error())
		return
	}

	code, ok := mappedData["httpstatuscode"].(float64)
	if !ok {
		s.logger.Error("invalid response format: httpstatuscode invalid")
		return
	}
	dataSeg, ok := mappedData["data"].(map[string]any)
	if !ok {
		s.logger.Error("invalid response format: data segment invalid")
		return
	}

	err := s.validator.Validate(ctx, &entity.SupplierOfferList{HTTPStatusCode: int(code), Data: &dataSeg})
	s.logger.Info(err.Error())
}
