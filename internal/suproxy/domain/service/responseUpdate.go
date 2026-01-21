package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// responseUpdateService implements the orchestration logic for updating persisted mock responses.
type responseUpdateService struct {
	logger           *slog.Logger
	tracer           trace.Tracer
	tagSearchService TagSearchService
	databaseRepo     repository.DatabaseRepository
	validatorService Validator
	cacheService     CacheService
}

// ResponseUpdateService defines the contract for updating stored responses based on a test request.
type ResponseUpdateService interface {
	UpdateResponse(ctx context.Context, mockRequest *entity.Request) error
}

// NewResponseUpdateService constructs a ResponseUpdateService with all required dependencies.
func NewResponseUpdateService(
	logger *slog.Logger,
	tracer trace.Tracer,
	tagSearchService TagSearchService,
	databaseRepo repository.DatabaseRepository,
	validatorService Validator,
	cacheService CacheService,
) (*responseUpdateService, error) {
	if err := assert.NotNil(logger, tracer, tagSearchService, databaseRepo, validatorService, cacheService); err != nil {
		return nil, fmt.Errorf("dependency cannot be nil, %w", err)
	}

	return &responseUpdateService{
		logger:           logger,
		tracer:           tracer,
		tagSearchService: tagSearchService,
		databaseRepo:     databaseRepo,
		validatorService: validatorService,
		cacheService:     cacheService,
	}, nil
}

// UpdateResponse is the public entrypoint.
func (s *responseUpdateService) UpdateResponse(
	ctx context.Context,
	mockRequest *entity.Request,
) error {
	ctx, span := s.tracer.Start(ctx, "ResponseUpdateService.UpdateResponse")
	defer span.End()

	if mockRequest == nil {
		return fmt.Errorf("mockRequest cannot be nil")
	}

	requestBody, err := parseRequestBody(mockRequest)
	if err != nil {
		return err
	}

	if err := validateRequestBody(requestBody); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// 1) Mock cache → early return
	done, err := s.handleMockCache(ctx, mockRequest)
	if err != nil {
		return err
	}
	if done {
		return nil
	}

	// 2) Resolve base entry (supplier cache OR parquet)
	baseEntry, err := s.resolveBaseEntry(ctx, mockRequest)
	if err != nil {
		return err
	}

	// 3) Run update pipeline
	return s.runUpdatePipeline(ctx, baseEntry, requestBody, mockRequest)
}

// handleMockCache checks the mock cache and returns true if processing is finished.
func (s *responseUpdateService) handleMockCache(
	ctx context.Context,
	mockRequest *entity.Request,
) (bool, error) {
	mockCached, mockHit, err := s.cacheService.Lookup(ctx, *mockRequest, true)
	if err != nil {
		return false, fmt.Errorf("mock cache lookup failed: %w", err)
	}

	if !mockHit {
		return false, nil
	}

	s.logger.Debug("update-response: mock cache hit")

	var items []json.RawMessage
	if err := json.Unmarshal(mockCached, &items); err != nil {
		return false, fmt.Errorf("failed to unmarshal mock cached response: %w", err)
	}

	if _, err := s.validatorService.Validate(
		ctx,
		&entity.SupplierResponse{
			Data: entity.SupplierOfferList{Items: items},
		},
		nil,
	); err != nil {
		return false, fmt.Errorf("mock cached response validation failed: %w", err)
	}

	return true, nil
}

// resolveBaseEntry determines the base response source.
func (s *responseUpdateService) resolveBaseEntry(
	ctx context.Context,
	mockRequest *entity.Request,
) (*entity.DatabaseEntry, error) {
	supplierCached, supplierHit, err := s.cacheService.Lookup(ctx, *mockRequest, false)
	if err != nil {
		return nil, fmt.Errorf("supplier cache lookup failed: %w", err)
	}

	if supplierHit {
		s.logger.Debug("update-response: supplier cache hit")

		return &entity.DatabaseEntry{
			Response: entity.Response{
				Response: string(supplierCached),
			},
			Tags: nil,
		}, nil
	}

	if mockRequest.Tags == "" {
		return nil, fmt.Errorf("update response requires non-empty tags")
	}

	parquetFiles, err := s.tagSearchService.FindKeysByTags(ctx, mockRequest.Tags)
	if err != nil {
		return nil, fmt.Errorf("tag-based search failed: %w", err)
	}

	if len(parquetFiles) == 0 {
		return nil, fmt.Errorf("no parquet files found for tags: %s", mockRequest.Tags)
	}

	requestTags := splitRequestTags(mockRequest.Tags)
	return s.selectBestMatchingResponse(ctx, parquetFiles, requestTags)
}

// runUpdatePipeline executes the update logic.
func (s *responseUpdateService) runUpdatePipeline(
	ctx context.Context,
	baseEntry *entity.DatabaseEntry,
	requestBody *entity.RequestBody,
	mockRequest *entity.Request,
) error {
	updatedEntry, err := updateResponseFields(baseEntry, requestBody)
	if err != nil {
		return fmt.Errorf("failed to update response fields: %w", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal([]byte(updatedEntry.Response.Response), &items); err != nil {
		return fmt.Errorf("failed to unmarshal updated response: %w", err)
	}

	if _, err := s.validatorService.Validate(
		ctx,
		&entity.SupplierResponse{
			Data: entity.SupplierOfferList{Items: items},
		},
		baseEntry.Tags,
	); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.databaseRepo.CreateRequest(ctx, *updatedEntry); err != nil {
		return fmt.Errorf("failed to save updated response: %w", err)
	}

	if err := s.cacheService.Store(
		ctx,
		*mockRequest,
		[]byte(updatedEntry.Response.Response),
		true,
		false,
	); err != nil {
		return fmt.Errorf("failed to store in cache: %w", err)
	}

	return nil
}

// updateResponseFields updates only time-dependent fields on the ODT response.
func updateResponseFields(
	dbEntry *entity.DatabaseEntry,
	requestBody *entity.RequestBody,
) (*entity.DatabaseEntry, error) {
	if dbEntry == nil || dbEntry.Response.Response == "" {
		return nil, fmt.Errorf("invalid database entry")
	}

	var response entity.ODTResponse
	if err := json.Unmarshal([]byte(dbEntry.Response.Response), &response); err != nil {
		return nil, fmt.Errorf("failed to parse ODT response: %w", err)
	}

	for i := range response.Data.Items {
		item := &response.Data.Items[i]

		item.DepartureDate = requestBody.DepartureDate
		item.ReturnDate = requestBody.ReturnDate
		item.CheckInHotel = requestBody.DepartureDate
		item.CheckOutHotel = requestBody.ReturnDate

		item.OvernightDuration =
			calculateOvernightDuration(
				requestBody.DepartureDate,
				requestBody.ReturnDate,
			)

		item.Availability = true

		if item.OfferID == "" {
			item.OfferID = uuid.NewString()
		}
	}

	serialized, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize updated response: %w", err)
	}

	updated := *dbEntry
	updated.Response.Response = string(serialized)
	updated.Updated = true

	return &updated, nil
}

// calculateOvernightDuration computes the number of nights between check-in and check-out dates.
func calculateOvernightDuration(checkIn, checkOut string) int {
	in, err1 := time.Parse("2006-01-02", checkIn)
	out, err2 := time.Parse("2006-01-02", checkOut)

	if err1 != nil || err2 != nil || !out.After(in) {
		return 0
	}

	return int(out.Sub(in).Hours() / 24)
}

// parseRequestBody parses the request body.
func parseRequestBody(mockRequest *entity.Request) (*entity.RequestBody, error) {
	if mockRequest == nil || mockRequest.Body == "" {
		return nil, fmt.Errorf("mockRequest body cannot be empty")
	}

	var body entity.RequestBody
	if err := json.Unmarshal([]byte(mockRequest.Body), &body); err != nil {
		return nil, fmt.Errorf("failed to parse mockRequest body: %w", err)
	}

	return &body, nil
}

// validateRequestBody validates required request fields.
func validateRequestBody(body *entity.RequestBody) error {
	if body.DepartureDate == "" || body.ReturnDate == "" {
		return fmt.Errorf("departureDate and returnDate are required")
	}
	if len(body.DepartureAirportList) == 0 {
		return fmt.Errorf("at least one departure airport is required")
	}
	if len(body.Travelers) == 0 {
		return fmt.Errorf("travelers must be provided")
	}
	return nil
}

// selectBestMatchingResponse chooses the stored response with the highest tag overlap.
func (s *responseUpdateService) selectBestMatchingResponse(
	ctx context.Context,
	keys []string,
	requestTags []string,
) (*entity.DatabaseEntry, error) {
	var bestEntry *entity.DatabaseEntry
	bestScore := -1

	for _, key := range keys {
		entry, err := s.databaseRepo.ReadRequest(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to read parquet file %s: %w", key, err)
		}

		score := tagScore(requestTags, splitResponseTags(entry.Tags))
		if score > bestScore {
			bestScore = score
			bestEntry = entry
		}
	}

	if bestEntry == nil {
		return nil, fmt.Errorf("no suitable response found for tags: %v", requestTags)
	}

	return bestEntry, nil
}

// splitRequestTags normalizes a comma-separated request tag string into a slice.
func splitRequestTags(tags string) []string {
	raw := strings.Split(tags, ",")
	out := make([]string, 0, len(raw))
	for _, t := range raw {
		if v := strings.TrimSpace(t); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// splitResponseTags extracts comparable tag names from a structured TagList.
func splitResponseTags(tags *sharedEntity.TagList) []string {
	if tags == nil {
		return nil
	}

	out := make([]string, 0, len(tags.Tags))
	for _, t := range tags.Tags {
		if t.Name != "" {
			out = append(out, t.Name)
		}
	}
	return out
}

// tagScore computes the number of shared tags between request and response.
func tagScore(requestTags, responseTags []string) int {
	set := make(map[string]struct{}, len(responseTags))
	for _, t := range responseTags {
		set[t] = struct{}{}
	}

	score := 0
	for _, t := range requestTags {
		if _, ok := set[t]; ok {
			score++
		}
	}
	return score
}
