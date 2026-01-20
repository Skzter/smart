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

type responseUpdateService struct {
	logger           *slog.Logger
	tracer           trace.Tracer
	tagSearchService TagSearchService
	databaseRepo     repository.DatabaseRepository
	validatorService Validator
}

// ResponseUpdateService defines the contract for updating persisted responses based on a given test request.
type ResponseUpdateService interface {
	UpdateResponse(ctx context.Context, mockRequest *entity.Request) error
}

// NewResponseUpdateService creates the service and ensures all required dependencies are provided.
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

// UpdateResponse orchestrates the full workflow: select, update, validate and persist a mock response.
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

	// Parse the request body as an update instruction for the response.
	updateRequest := &entity.UpdateResponse{}
	if err := json.Unmarshal([]byte(mockRequest.Body), updateRequest); err != nil {
		return fmt.Errorf("failed to parse mockRequest body as UpdateResponse: %w", err)
	}

	if mockRequest.Tags == "" {
		return fmt.Errorf("update response requires non-empty tags")
	}

	// Find stored responses matching the request tags.
	parquetFiles, err := s.tagSearchService.FindKeysByTags(ctx, mockRequest.Tags)
	if err != nil {
		return fmt.Errorf("tag-based search failed: %w", err)
	}

	if len(parquetFiles) == 0 {
		return fmt.Errorf("no parquet files found for tags: %s", mockRequest.Tags)
	}

	requestTags := splitRequestTags(mockRequest.Tags)

	// Select the best matching stored response based on tag overlap.
	dbEntry, err := s.selectBestMatchingResponse(ctx, parquetFiles, requestTags)
	if err != nil {
		return err
	}

	// Parse the same request body as a test-request context (dates, travelers, etc.).
	requestBody, err := parseRequestBody(mockRequest)
	if err != nil {
		return fmt.Errorf("failed to parse request body: %w", err)
	}

	// Apply rule-based transformations to the stored response.
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

	// Persist the updated response back to storage.
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

// splitRequestTags normalizes a comma-separated tag string into a clean slice.
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

// tagScore counts how many tags are shared between request and response.
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

// updateResponseFields applies deterministic, rule-based updates to a stored response.
func updateResponseFields(
	dbEntry *entity.DatabaseEntry,
	updateRequest *entity.UpdateResponse,
	requestBody *entity.RequestBody,
) (*entity.DatabaseEntry, error) {
	if dbEntry == nil || dbEntry.Response.Response == "" {
		return nil, fmt.Errorf("invalid database entry")
	}

	// Deserialize the stored response into an editable structure.
	var response entity.UpdateResponse
	if err := json.Unmarshal([]byte(dbEntry.Response.Response), &response); err != nil {
		return nil, fmt.Errorf("failed to parse stored response JSON: %w", err)
	}

	// Iterate over items and apply transformation rules field by field.
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

		if upd.Price > 0 {
			item.Price = upd.Price
		} else if updateRequest.Data.MinPrice > 0 {
			item.Price *= updateRequest.Data.MinPrice
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

	// Serialize the updated response back to JSON for persistence.
	serialized, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize updated response: %w", err)
	}

	updated := *dbEntry
	updated.Response.Response = string(serialized)
	updated.Updated = true

	return &updated, nil
}

// calculateOvernightDuration computes the number of nights between check-in and check-out.
func calculateOvernightDuration(checkIn, checkOut string) int {
	in, err1 := time.Parse("2006-01-02", checkIn)
	out, err2 := time.Parse("2006-01-02", checkOut)

	if err1 != nil || err2 != nil || !out.After(in) {
		return 0
	}

	return int(out.Sub(in).Hours() / 24)
}

// parseRequestBody parses the test-request payload into a typed RequestBody structure.
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
