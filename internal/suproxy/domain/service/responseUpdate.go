package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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
		return fmt.Errorf("failed to parse Requestbody, %w", err)
	}

	if err := validateRequestBody(requestBody); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// 1) Mock cache → early return
	done, err := s.handleMockCache(ctx, mockRequest)
	if err != nil {
		return fmt.Errorf("failed to handleMockCache, %w", err)
	}
	if done {
		return nil
	}

	// 2) Resolve base entry (supplier cache OR parquet)
	baseEntry, err := s.resolveBaseEntry(ctx, mockRequest)
	if err != nil {
		return fmt.Errorf("failed to resolve supplier or cache baseentry, %w", err)
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
	updatedEntry, relevantItems, err := updateResponseFields(baseEntry, requestBody)
	if err != nil {
		return fmt.Errorf("failed to update response fields: %w", err)
	}

	if err := runDeterministicValidation(relevantItems, requestBody); err != nil {
		return fmt.Errorf("deterministic validation failed: %w", err)
	}

	validationItems, err := marshalOffersForValidation(relevantItems)
	if err != nil {
		return fmt.Errorf("failed to marshal offers for validation: %w", err)
	}

	statusCode := extractHTTPStatusCode(updatedEntry.Response.Response)
	if _, err := s.validatorService.Validate(
		ctx,
		&entity.SupplierResponse{
			HTTPStatusCode: statusCode,
			Data:           entity.SupplierOfferList{Items: validationItems},
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
) (*entity.DatabaseEntry, []*entity.ODTItem, error) {
	if dbEntry == nil || dbEntry.Response.Response == "" {
		return nil, nil, fmt.Errorf("invalid database entry")
	}

	var response entity.ODTResponse
	if err := json.Unmarshal([]byte(dbEntry.Response.Response), &response); err != nil {
		return nil, nil, fmt.Errorf("failed to parse ODT response: %w", err)
	}

	checkIn, checkOut, err := normalizeRequestDates(requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid request dates: %w", err)
	}

	relevantItems, err := selectRelevantOffers(&response, requestBody)
	if err != nil {
		return nil, nil, err
	}

	for _, item := range relevantItems {
		updateOfferDates(item, checkIn, checkOut)
		if item.OfferID == "" {
			item.OfferID = uuid.NewString()
		}
	}

	serialized, err := json.Marshal(response)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize updated response: %w", err)
	}

	updated := *dbEntry
	updated.Response.Response = string(serialized)
	updated.Updated = true

	return &updated, relevantItems, nil
}

// calculateOvernightDuration computes the number of nights between check-in and check-out dates.
func calculateOvernightDuration(checkIn, checkOut time.Time) int {
	if checkOut.Before(checkIn) {
		return 0
	}

	days := int(checkOut.Sub(checkIn).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

const odtTimestampLayout = "2006-01-02T15:04:05-0700"

func normalizeRequestDates(body *entity.RequestBody) (time.Time, time.Time, error) {
	if body == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("request body must be provided")
	}

	if body.DepartureDate == "" || body.ReturnDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("departureDate and returnDate must be set")
	}

	checkIn, err := time.Parse("2006-01-02", body.DepartureDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("departureDate is invalid: %w", err)
	}
	checkOut, err := time.Parse("2006-01-02", body.ReturnDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("returnDate is invalid: %w", err)
	}
	if !checkOut.After(checkIn) {
		return time.Time{}, time.Time{}, fmt.Errorf("returnDate must be after departureDate")
	}

	return checkIn, checkOut, nil
}

func formatODTTimestamp(t time.Time) string {
	return t.UTC().Format(odtTimestampLayout)
}

func selectRelevantOffers(response *entity.ODTResponse, body *entity.RequestBody) ([]*entity.ODTItem, error) {
	if response == nil || len(response.Data.Items) == 0 {
		return nil, fmt.Errorf("response contains no offers")
	}

	airportSet := buildAirportSet(body.DepartureAirportList)
	filtered := make([]*entity.ODTItem, 0, len(response.Data.Items))

	for i := range response.Data.Items {
		item := &response.Data.Items[i]
		if !matchesTravelType(item, body) {
			continue
		}
		if len(airportSet) > 0 && !offersAirportMatch(item, airportSet) {
			continue
		}
		filtered = append(filtered, item)
	}

	if len(filtered) == 0 {
		if len(airportSet) > 0 || body.TravelType != "" {
			return nil, fmt.Errorf("no offers match the requested filters (airports=%v travelType=%s)", body.DepartureAirportList, body.TravelType)
		}
		filtered = make([]*entity.ODTItem, 0, len(response.Data.Items))
		for i := range response.Data.Items {
			filtered = append(filtered, &response.Data.Items[i])
		}
	}

	return filtered, nil
}

func buildAirportSet(list []string) map[string]struct{} {
	out := make(map[string]struct{}, len(list))
	for _, raw := range list {
		if code := strings.ToUpper(strings.TrimSpace(raw)); code != "" {
			out[code] = struct{}{}
		}
	}
	return out
}

func matchesTravelType(item *entity.ODTItem, body *entity.RequestBody) bool {
	if body == nil || body.TravelType == "" {
		return true
	}
	return strings.EqualFold(body.TravelType, item.TravelType)
}

func offersAirportMatch(item *entity.ODTItem, airports map[string]struct{}) bool {
	if item == nil {
		return false
	}
	code := strings.ToUpper(strings.TrimSpace(item.Flight.OutboundDepartureAirport.Code))
	if code == "" {
		return false
	}
	_, ok := airports[code]
	return ok
}

func updateOfferDates(item *entity.ODTItem, checkIn, checkOut time.Time) {
	if item == nil {
		return
	}

	checkInTs := formatODTTimestamp(checkIn)
	checkOutTs := formatODTTimestamp(checkOut)
	item.DepartureDate = checkInTs
	item.ReturnDate = checkOutTs
	item.Accommodation.CheckInDate = checkInTs
	item.Accommodation.CheckOutDate = checkOutTs
	item.OvernightDuration.CheckInHotel = checkInTs
	item.OvernightDuration.CheckOutHotel = checkOutTs
	item.OvernightDuration.NightsInHotel = calculateOvernightDuration(checkIn, checkOut)
}

func runDeterministicValidation(items []*entity.ODTItem, body *entity.RequestBody) error {
	if len(items) == 0 {
		return fmt.Errorf("no offers to validate")
	}

	var errs []string
	for _, item := range items {
		if err := validateMandatoryFields(item); err != nil {
			errs = append(errs, fmt.Sprintf("offer %s: %v", item.OfferID, err))
			continue
		}
		if err := validateDateConsistency(item, body); err != nil {
			errs = append(errs, fmt.Sprintf("offer %s: %v", item.OfferID, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

func marshalOffersForValidation(items []*entity.ODTItem) ([]json.RawMessage, error) {
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	return out, nil
}

func extractHTTPStatusCode(payload string) int {
	var meta struct {
		HTTPStatusCode int `json:"httpstatuscode"`
	}
	if err := json.Unmarshal([]byte(payload), &meta); err == nil && meta.HTTPStatusCode != 0 {
		return meta.HTTPStatusCode
	}
	return http.StatusOK
}

func validateMandatoryFields(item *entity.ODTItem) error {
	if item == nil {
		return fmt.Errorf("offer is nil")
	}
	switch {
	case item.OfferID == "":
		return fmt.Errorf("offerid is missing")
	case item.DepartureDate == "":
		return fmt.Errorf("departuredate is missing")
	case item.ReturnDate == "":
		return fmt.Errorf("returndate is missing")
	case item.Accommodation.CheckInDate == "":
		return fmt.Errorf("accommodation.checkindate is missing")
	case item.Accommodation.CheckOutDate == "":
		return fmt.Errorf("accommodation.checkoutdate is missing")
	case len(item.Accommodation.Rooms) == 0:
		return fmt.Errorf("accommodation.rooms is empty")
	case item.Flight.OutboundDepartureAirport.Code == "":
		return fmt.Errorf("outbound departure airport is missing")
	case item.Flight.InboundDepartureAirport.Code == "":
		return fmt.Errorf("inbound departure airport is missing")
	}
	return nil
}

func validateDateConsistency(item *entity.ODTItem, body *entity.RequestBody) error {
	dep, err := parseFlexibleTimestamp(item.DepartureDate)
	if err != nil {
		return fmt.Errorf("departureDate invalid: %w", err)
	}
	ret, err := parseFlexibleTimestamp(item.ReturnDate)
	if err != nil {
		return fmt.Errorf("returnDate invalid: %w", err)
	}
	checkIn, err := parseFlexibleTimestamp(item.Accommodation.CheckInDate)
	if err != nil {
		return fmt.Errorf("accommodation.checkInDate invalid: %w", err)
	}
	checkOut, err := parseFlexibleTimestamp(item.Accommodation.CheckOutDate)
	if err != nil {
		return fmt.Errorf("accommodation.checkOutDate invalid: %w", err)
	}
	if !ret.After(dep) {
		return fmt.Errorf("returnDate must be after departureDate")
	}
	if !checkOut.After(checkIn) {
		return fmt.Errorf("accommodation checkout must be after checkin")
	}
	nights := calculateOvernightDuration(checkIn, checkOut)
	if item.OvernightDuration.NightsInHotel != nights {
		return fmt.Errorf("overnight duration %d does not match stay length %d", item.OvernightDuration.NightsInHotel, nights)
	}
	if !checkIn.Equal(dep) {
		return fmt.Errorf("accommodation.checkInDate and departureDate diverged")
	}
	if !checkOut.Equal(ret) {
		return fmt.Errorf("accommodation.checkOutDate and returnDate diverged")
	}
	if body != nil && body.TravelType != "" && !strings.EqualFold(body.TravelType, item.TravelType) {
		return fmt.Errorf("offer travelType %q differs from request %q", item.TravelType, body.TravelType)
	}
	return nil
}

func parseFlexibleTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("timestamp is empty")
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(odtTimestampLayout, value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("timestamp %q is not RFC3339 or %s", value, odtTimestampLayout)
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
