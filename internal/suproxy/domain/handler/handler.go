package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// SuproxyController handles the HTTP requests for the Suproxy service
type SuproxyController struct {
	logger          *slog.Logger
	config          *config.Suproxy
	client          *http.Client
	validator       service.Validator
	db              service.DatabaseService
	tracer          trace.Tracer
	syncer          service.TaglistSync
	metrics         sharedService.MetricsService
	cache           service.CacheService
	responseUpdater service.ResponseUpdateService
}

// NewSuproxyController creates a new instance of SuproxyController
func NewSuproxyController(
	logger *slog.Logger,
	config *config.Suproxy,
	val service.Validator,
	client *http.Client,
	db service.DatabaseService,
	tracer trace.Tracer,
	syncer service.TaglistSync,
	metrics sharedService.MetricsService,
	cache service.CacheService,
	responseUpdater service.ResponseUpdateService,
) (*SuproxyController, error) {
	if err := assert.NotNil(logger, config, val, client, db, tracer, syncer, metrics, cache, responseUpdater); err != nil {
		return nil, err
	}

	return &SuproxyController{
		logger:          logger,
		config:          config,
		client:          client,
		validator:       val,
		db:              db,
		tracer:          tracer,
		syncer:          syncer,
		metrics:         metrics,
		cache:           cache,
		responseUpdater: responseUpdater,
	}, nil
}

// PostOfferlist handles the POST request to the /api/v1/Offerlist endpoint
func (s *SuproxyController) PostOfferlist(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var request entity.Request

	ctx, span := s.tracer.Start(ctx, "suproxyController.PostOfferlist")
	defer span.End()

	if err := c.BindJSON(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		s.logger.Error("Failed to bind JSON", "error", err)
		s.metrics.IncRequestError("invalid_json")
		s.metrics.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.Tags != "" {
		if err := s.responseUpdater.UpdateResponse(ctx, &request); err != nil {
			// Failure to update the response is logged but not returned as an error, since it is non-critical and should not block request processing.
			span.RecordError(err)
			s.logger.Error("response update failed", "error", err)
		}
	}

	if cached, hit, err := s.cache.Lookup(ctx, request, false); err == nil && hit {
		s.logger.Debug("cache: handler served response from cache")
		c.Data(http.StatusOK, "application/json", cached)
		return
	}

	body, code, err := s.fetchOffers(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch offers")
		s.logger.Error("Failed to fetch offers", "error", err)
		s.metrics.IncRequestError("fetch_failed")
		s.metrics.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	s.metrics.RecordStatusCode(code)

	if code == http.StatusOK {
		go s.HandleRequest(c.Copy(), request, body)
		span.SetStatus(codes.Ok, "supplier request successful")
		s.metrics.IncRequestSuccess()
	} else {
		span.SetStatus(codes.Error, "supplier request failed")
		s.logger.Error("supplier request failed", "code", code)
		s.metrics.IncRequestError("supplier_error")
	}

	isError := code != http.StatusOK
	if err := s.cache.Store(ctx, request, *body, false, isError); err != nil {
		s.logger.Error("cache: failed to store response", "error", err)
	}

	s.metrics.RecordRequestDuration(time.Since(start))
	c.Data(code, "application/json", *body)
}

func (s *SuproxyController) fetchOffers(ctx context.Context, request entity.Request) (*[]byte, int, error) {
	ctx, span := s.tracer.Start(ctx, "suproxyController.fetchOffers")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", http.MethodPost),
		attribute.String("http.url", request.Destination),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, request.Destination, bytes.NewBuffer([]byte(request.Body)))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create HTTP request")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "HTTP request failed")
		return nil, 0, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	span.SetAttributes(
		attribute.Int("http.status_code", resp.StatusCode),
	)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read response body")
		return nil, 0, err
	}

	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, "HTTP request returned error status")
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return &body, resp.StatusCode, nil
}

// HandleRequest validates given response data and stores it combined with request
func (s *SuproxyController) HandleRequest(ctx context.Context, req entity.Request, respData *[]byte) {
	var list entity.SupplierResponse

	ctx, span := s.tracer.Start(ctx, "suproxyController.HandleRequest")
	defer span.End()

	if err := json.Unmarshal(*respData, &list); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshal response data")
		s.logger.Error(err.Error())
		return
	}
	tags, err := s.validator.Validate(ctx, &list, s.syncer.GetCurrentTaglist())
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	err = s.syncer.SyncTaglist(ctx, tags)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to validate response data")
		s.logger.Error(err.Error())
		return
	}

	if err := s.store(ctx, &req, &list, tags); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store response data")
		s.logger.Error(err.Error())
		return
	}

	span.SetStatus(codes.Ok, "response data stored successfully")
}

func (s *SuproxyController) store(ctx context.Context, req *entity.Request, resp *entity.SupplierResponse, tags *sharedEntity.TagList) error {
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
