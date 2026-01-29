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
	cacheService CacheService,
) (ResponseUpdateService, error) {
	if err := assert.NotNil(logger, tracer, tagSearchService, databaseRepo, cacheService); err != nil {
		return nil, fmt.Errorf("dependency cannot be nil, %w", err)
	}

	return &responseUpdateService{
		logger:           logger,
		tracer:           tracer,
		tagSearchService: tagSearchService,
		databaseRepo:     databaseRepo,
		cacheService:     cacheService,
	}, nil
}

// UpdateResponse is the public entrypoint.
func (s *responseUpdateService) UpdateResponse(
	ctx context.Context,
	mockRequest *entity.Request,
) error {
	err := assert.NotNil(ctx, mockRequest)
	if err != nil {
		return fmt.Errorf("mockRequest and context cannot be nil, %w", err)
	}

	ctx, span := s.tracer.Start(ctx, "ResponseUpdateService.UpdateResponse")
	defer span.End()

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
		s.logger.Warn(
			"mock cache error, treating as cache miss",
			"error", err,
		)
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

	if !json.Valid(mockCached) {
		return false, fmt.Errorf("mock cached response is not valid json")
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

	statusCode := extractHTTPStatusCode(updatedEntry.Response.Response)
	if statusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code: %d", statusCode)
	}

	updatedEntry.Request = *mockRequest
	updatedEntry.Tags = buildTagListFromString(mockRequest.Tags)

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

func buildTagListFromString(tags string) *sharedEntity.TagList {
	if tags == "" {
		return nil
	}

	raw := strings.Split(tags, ",")
	list := make([]sharedEntity.Tag, 0, len(raw))

	for _, t := range raw {
		if name := strings.TrimSpace(t); name != "" {
			list = append(list, sharedEntity.Tag{Name: name})
		}
	}

	if len(list) == 0 {
		return nil
	}

	return &sharedEntity.TagList{Tags: list}
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
			item.OfferID = entity.FlexibleString(uuid.NewString())
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
	hasFilters := len(airportSet) > 0 || body.TravelType != ""

	if !hasFilters {
		result := make([]*entity.ODTItem, 0, len(response.Data.Items))
		for i := range response.Data.Items {
			result = append(result, &response.Data.Items[i])
		}
		return result, nil
	}

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
		return nil, fmt.Errorf(
			"no offers match the requested filters (airports=%v travelType=%s)",
			body.DepartureAirportList,
			body.TravelType,
		)
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

	for _, item := range items {
		if err := validateODTItem(item, body); err != nil {
			return fmt.Errorf("validation error: %w", err)
		}
	}
	return nil
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

func validateODTItem(item *entity.ODTItem, body *entity.RequestBody) error {
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
		return fmt.Errorf(
			"overnight duration %d does not match stay length %d",
			item.OvernightDuration.NightsInHotel,
			nights,
		)
	}

	if !checkIn.Equal(dep) {
		return fmt.Errorf("accommodation.checkInDate and departureDate diverged")
	}

	if !checkOut.Equal(ret) {
		return fmt.Errorf("accommodation.checkOutDate and returnDate diverged")
	}

	if body != nil && body.TravelType != "" &&
		!strings.EqualFold(body.TravelType, item.TravelType) {
		return fmt.Errorf(
			"offer travelType %q differs from request %q",
			item.TravelType,
			body.TravelType,
		)
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

	var payload entity.UpdateRequestPayload
	if err := json.Unmarshal([]byte(mockRequest.Body), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse request body: %w", err)
	}

	if len(payload.Params) == 0 {
		return nil, fmt.Errorf("params must contain at least one element")
	}

	return &payload.Params[0], nil
}

// validateRequestBody validates required request fields.
func validateRequestBody(body *entity.RequestBody) error {
	if body.DepartureDate == "" || body.ReturnDate == "" {
		return fmt.Errorf("departureDate and returnDate are required")
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
	requestTagSet := buildTagSet(requestTags)

	var bestEntry *entity.DatabaseEntry
	bestScore := -1

	for _, key := range keys {
		entry, err := s.databaseRepo.ReadRequest(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to read parquet file %s: %w", key, err)
		}

		score := tagScore(requestTagSet, splitResponseTags(entry.Tags))
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

func buildTagSet(tags []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		set[t] = struct{}{}
	}
	return set
}

// tagScore computes the number of shared tags between request and response.
func tagScore(requestTagSet map[string]struct{}, responseTags []string) int {
	score := 0
	for _, t := range responseTags {
		if _, ok := requestTagSet[t]; ok {
			score++
		}
	}
	return score
}
