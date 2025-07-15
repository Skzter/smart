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
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// SuproxyController handles the HTTP requests for the Suproxy service
type SuproxyController struct {
	logger    *slog.Logger
	config    *config.Config
	client    *http.Client
	validator service.Validator
	db        service.DatabaseService
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(logger *slog.Logger, config *config.Config, val service.Validator, client *http.Client, db service.DatabaseService) (*SuproxyController, error) {
	if err := assert.NotNil(logger, config, val, client, db); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:    logger,
		config:    config,
		client:    client,
		validator: val,
		db:        db,
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
		go s.HandleRequest(c.Copy(), &request, body)
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
			panic(err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return &body, resp.StatusCode, nil
}

func (s *SuproxyController) HandleRequest(ctx context.Context, req *entity.Request, respData *[]byte) {
	var list entity.SupplierResponse
	if err := json.Unmarshal(*respData, &list); err != nil {
		s.logger.Error(err.Error())
		return
	}

	tags, err := s.validator.Validate(ctx, &list)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	if err := s.store(ctx, req, &list, tags); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *SuproxyController) store(ctx context.Context, req *entity.Request, resp *entity.SupplierResponse, tags *[]string) error {
	mresp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	dbentry := entity.DatabaseEntry{
		Request: *req,
		Response: entity.Response{
			Response: string(mresp),
		},
		Tags: *tags,
	}

	return s.db.SaveDbEntry(ctx, dbentry)
}
