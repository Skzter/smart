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
) (*responseUpdateService, error) {
	if err := assert.NotNil(logger, tracer, tagSearchService, databaseRepo); err != nil {
		return nil, fmt.Errorf("dependency cannot be nil, %w", err)
	}

	return &responseUpdateService{
		logger:           logger,
		tracer:           tracer,
		tagSearchService: tagSearchService,
		databaseRepo:     databaseRepo,
		validatorService: validatorService,
	}, nil
}

// UpdateResponse selects, updates, validates, and persists a stored mock response deterministically.
func (s *responseUpdateService) UpdateResponse(
	ctx context.Context,
	mockRequest *entity.Request,
) error {
	ctx, span := s.tracer.Start(ctx, "ResponseUpdateService.UpdateResponse")
	defer span.End()

	if mockRequest == nil {
		return fmt.Errorf("mockRequest cannot be nil")
	}

	if mockRequest.Body == "" {
		return fmt.Errorf("mockRequest body cannot be empty")
	}

	updateRequest := &entity.UpdateResponse{}
	if err := json.Unmarshal([]byte(mockRequest.Body), updateRequest); err != nil {
		return fmt.Errorf("failed to parse mockRequest body as UpdateResponse: %w", err)
	}

	if mockRequest.Tags == "" {
		return fmt.Errorf("update response requires non-empty tags")
	}

	parquetFiles, err := s.tagSearchService.FindKeysByTags(ctx, mockRequest.Tags)
	if err != nil {
		return fmt.Errorf("tag-based search failed: %w", err)
	}

	if len(parquetFiles) == 0 {
		return fmt.Errorf("no parquet files found for tags: %s", mockRequest.Tags)
	}

	requestTags := splitRequestTags(mockRequest.Tags)

	dbEntry, err := s.selectBestMatchingResponse(ctx, parquetFiles, requestTags)
	if err != nil {
		return err
	}

	requestBody, err := parseRequestBody(mockRequest)
	if err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	updatedEntry, err := updateResponseFields(dbEntry, updateRequest, requestBody)
	if err != nil {
		return fmt.Errorf("failed to update response fields: %w", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal([]byte(updatedEntry.Response.Response), &items); err != nil {
		return fmt.Errorf("failed to unmarshal updated response: %w", err)
	}

	if _, err := s.validatorService.Validate(
		ctx,
		&entity.SupplierResponse{Data: entity.SupplierOfferList{Items: items}},
		dbEntry.Tags,
	); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.databaseRepo.CreateRequest(ctx, *updatedEntry); err != nil {
		return fmt.Errorf("failed to save updated response: %w", err)
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

// updateResponseFields applies rule-based item updates and recomputes all ODT data-level aggregates.
func updateResponseFields(
	dbEntry *entity.DatabaseEntry,
	updateRequest *entity.UpdateResponse,
	requestBody *entity.RequestBody,
) (*entity.DatabaseEntry, error) {
	if dbEntry == nil || dbEntry.Response.Response == "" {
		return nil, fmt.Errorf("invalid database entry")
	}

	var response entity.UpdateResponse
	if err := json.Unmarshal([]byte(dbEntry.Response.Response), &response); err != nil {
		return nil, fmt.Errorf("failed to parse stored response JSON: %w", err)
	}

	minPrice := 0.0
	maxPrice := 0.0
	availableOffers := 0

	minLen := len(response.Data.Items)
	if len(updateRequest.Data.Items) < minLen {
		minLen = len(updateRequest.Data.Items)
	}

	for i := 0; i < minLen; i++ {
		item := &response.Data.Items[i]
		upd := &updateRequest.Data.Items[i]

		if upd.DepartureDate != "" {
			item.DepartureDate = upd.DepartureDate
		} else {
			item.DepartureDate = requestBody.DepartureDate
		}

		if upd.ReturnDate != "" {
			item.ReturnDate = upd.ReturnDate
		} else {
			item.ReturnDate = requestBody.ReturnDate
		}

		if upd.CheckInHotel != "" {
			item.CheckInHotel = upd.CheckInHotel
		}

		if upd.CheckOutHotel != "" {
			item.CheckOutHotel = upd.CheckOutHotel
		}

		item.OvernightDuration = calculateOvernightDuration(
			item.CheckInHotel,
			item.CheckOutHotel,
		)

		item.Availability = true
		availableOffers++

		if upd.Price > 0 {
			item.Price = upd.Price
		} else if updateRequest.Data.MinPrice > 0 {
			item.Price *= updateRequest.Data.MinPrice
		}

		if i == 0 || item.Price < minPrice {
			minPrice = item.Price
		}
		if item.Price > maxPrice {
			maxPrice = item.Price
		}

		if upd.Currency != "" {
			item.Currency = upd.Currency
		}
		if upd.Description != "" {
			item.Description = upd.Description
		}

		if upd.OfferID != "" {
			item.OfferID = upd.OfferID
		} else {
			item.OfferID = uuid.NewString()
		}
	}

	response.Data.ResultCount = len(response.Data.Items)
	response.Data.CalculatedResultCount = len(response.Data.Items)
	response.Data.AvailableOffers = availableOffers
	response.Data.MinPrice = minPrice
	response.Data.MaxPrice = maxPrice

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

// parseRequestBody parses the raw request body into a typed RequestBody structure.
func parseRequestBody(mockRequest *entity.Request) (*entity.RequestBody, error) {
	if mockRequest == nil {
		return nil, fmt.Errorf("mockRequest cannot be nil")
	}

	if mockRequest.Body == "" {
		return nil, fmt.Errorf("mockRequest body cannot be empty")
	}

	var body entity.RequestBody
	if err := json.Unmarshal([]byte(mockRequest.Body), &body); err != nil {
		return nil, fmt.Errorf("failed to parse mockRequest body: %w", err)
	}

	return &body, nil
}
