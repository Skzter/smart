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
	if err := assert.NotNil(logger, config); err != nil {
		return nil, err
	}

	val, err := validator.NewValidator(logger, config)
	if err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:    logger,
		config:    config,
		client:    &http.Client{},
		validator: val,
	}, nil
}

// PostOfferlist handles the POST request to the /api/v1/Offerlist endpoint
func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	var request entity.Request

	if err := c.BindJSON(&request); err != nil {
		s.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	body, code, err := s.fetchOffers(request)
	if err != nil {
		s.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if code == http.StatusOK {
		done := make(chan any)
		go s.handleRequest(c, &request, body, done)
		defer func() { <-done }()
	} else {
		s.logger.Error("supplier request failed", "code", code)
	}

	c.Data(code, "application/json", *body)
}

func (s *SuproxyController) fetchOffers(request entity.Request) (*[]byte, int, error) {
	req, err := http.NewRequest(http.MethodPost, request.Destination, bytes.NewBuffer([]byte(request.Request)))
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		if err := req.Body.Close(); err != nil {
			panic(err)
		}
	}()

	for key, value := range request.Header {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send request", "error", err)
		return nil, 0, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Failed to close response body", "error", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return &body, resp.StatusCode, nil
}

func (s *SuproxyController) handleRequest(ctx context.Context, req *entity.Request, respData *[]byte, done chan<- any) {
	defer close(done)

	var list entity.SupplierResponse
	if err := json.Unmarshal(*respData, &list); err != nil {
		s.logger.Error(err.Error())
		return
	}

	if err := s.validator.Validate(ctx, &list); err != nil {
		s.logger.Error(err.Error())
		return
	}

	tags := []string{}

	if err := s.store(req, &list, tags); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

//nolint:unparam
func (s *SuproxyController) store(req *entity.Request, resp *entity.SupplierResponse, tags []string) error {
	// speichern

	return nil
}

// setupRouter initializes the Gin router and sets up the routes for the API
func SetupRouter(h *SuproxyController) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	return router
}
