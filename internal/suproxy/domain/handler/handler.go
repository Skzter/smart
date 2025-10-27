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
	logger      *slog.Logger
	config      *config.Config
	client      *http.Client
	validator   service.Validator
	db          service.DatabaseService
	tagSearch   service.TagSearchService
	handleAsync bool
}

// NewSuproxyController creates a new instance of SuproxyController
//
//nolint:lll
func NewSuproxyController(
	logger *slog.Logger,
	config *config.Config,
	val service.Validator,
	client *http.Client,
	db service.DatabaseService,
	tagSearch service.TagSearchService,
) (*SuproxyController, error) {
	if err := assert.NotNil(logger, config, val, client, db); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:      logger,
		config:      config,
		client:      client,
		validator:   val,
		db:          db,
		tagSearch:   tagSearch,
		handleAsync: true,
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

	if request.Prompt != "" {
		matchingKeys, err := s.tagSearch.FindKeysByTag(c.Request.Context(), request.Prompt)
		switch {
		case err != nil:
			s.logger.Error("Tag-based search failed", "error", err)
		case len(matchingKeys) == 0:
			s.logger.Error("No keys found in prompt", "error", err)
		default:
			s.logger.Info("Matching keys found", "keys", matchingKeys)
		}
	}

	body, code, err := s.fetchOffers(request)
	if err != nil {
		s.logger.Error("Failed to fetch offers", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if code == http.StatusOK {
		// Pass a context.Context to HandleRequest. Use a copy of the Gin
		// context to safely access the underlying *http.Request when
		// running in a background goroutine. Allow synchronous handling
		// when configured.
		if s.handleAsync {
			go s.HandleRequest(c.Copy(), request, body)
		} else {
			s.HandleRequest(c, request, body)
		}
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

// HandleRequest validates given response data and stores it combined with request
func (s *SuproxyController) HandleRequest(ctx context.Context, req entity.Request, respData *[]byte) {
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

	if err := s.store(ctx, &req, &list, tags); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *SuproxyController) store(ctx context.Context, req *entity.Request, resp *entity.SupplierResponse, tags []string) error {
	mresp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	dbentry := entity.DatabaseEntry{
		Request: *req,
		Response: entity.Response{
			Response: string(mresp),
		},
		Tags: tags,
	}

	return s.db.SaveDbEntry(ctx, dbentry)
}

// SetHandleAsync sets whether HandleRequest should run asynchronously.
// Use false in tests to make behavior deterministic.
func (s *SuproxyController) SetHandleAsync(on bool) {
	s.handleAsync = on
}
