package service

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	mockRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks/repository"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// TestNewResponseUpdateService does verify constructor behavior and dependency validation.
func TestNewResponseUpdateService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		tracer  trace.Tracer
		cache   CacheService
		tags    TagSearchService
		repo    repository.DatabaseRepository
		wantErr bool
	}{
		{
			name:    "all dependencies provided",
			logger:  logger,
			tracer:  tracer,
			cache:   mocks.NewMockCacheService(t),
			tags:    mocks.NewMockTagSearchService(t),
			repo:    mockRepo.NewMockDatabaseRepository(t),
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			wantErr: true,
		},
		{
			name:    "nil tracer",
			logger:  logger,
			tracer:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewResponseUpdateService(
				tt.logger,
				tt.tracer,
				tt.tags,
				tt.repo,
				tt.cache,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

// TestSelectRelevantOffers does validate offer filtering based on request criteria.
func TestSelectRelevantOffers(t *testing.T) {
	tests := []struct {
		name      string
		response  entity.ODTResponse
		body      *entity.RequestBody
		wantCount int
		wantErr   bool
	}{
		{
			name: "no filters → all offers",
			response: buildResponse(
				sampleItem("1", "package", "FRA"),
				sampleItem("2", "package", "MUC"),
			),
			body:      &entity.RequestBody{},
			wantCount: 2,
		},
		{
			name: "airport filter",
			response: buildResponse(
				sampleItem("1", "package", "FRA"),
				sampleItem("2", "package", "MUC"),
			),
			body: &entity.RequestBody{
				DepartureAirportList: []string{"FRA"},
			},
			wantCount: 1,
		},
		{
			name: "travelType filter",
			response: buildResponse(
				sampleItem("1", "package", "FRA"),
				sampleItem("2", "flight", "FRA"),
			),
			body:      &entity.RequestBody{TravelType: "flight"},
			wantCount: 1,
		},
		{
			name: "filters exclude all → error",
			response: buildResponse(
				sampleItem("1", "package", "FRA"),
			),
			body: &entity.RequestBody{
				DepartureAirportList: []string{"AMS"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := selectRelevantOffers(&tt.response, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, items, tt.wantCount)
		})
	}
}

// nolint:funlen
// TestUpdateResponse does validate the full UpdateResponse workflow including cache, tags, validation, and persistence.
func TestUpdateResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		tags      string
		body      entity.RequestBody
		setupMock func(
			*mocks.MockCacheService,
			*mocks.MockTagSearchService,
			*mockRepo.MockDatabaseRepository,
		)
		wantErr bool
	}{
		{
			name: "mock cache hit → short circuit",
			body: sampleRequestBody("package", []string{"FRA"}),
			setupMock: func(
				cache *mocks.MockCacheService,
				_ *mocks.MockTagSearchService,
				_ *mockRepo.MockDatabaseRepository,
			) {
				// 1) mock cache HIT
				cache.On(
					"Lookup",
					mock.Anything,
					mock.Anything,
					true,
				).Return([]byte(`[]`), true, nil).Once()

				// 2) supplier lookup may still happen depending on flow
				cache.On(
					"Lookup",
					mock.Anything,
					mock.Anything,
					false,
				).Return(nil, false, nil).Maybe()
			},
			wantErr: false,
		},
		{
			name: "tag based lookup → update + persist",
			tags: "tag-a",
			body: sampleRequestBody("package", []string{"FRA"}),
			setupMock: func(
				cache *mocks.MockCacheService,
				tags *mocks.MockTagSearchService,
				repo *mockRepo.MockDatabaseRepository,
			) {
				cache.On("Lookup", mock.Anything, mock.Anything, true).
					Return(nil, false, nil).Once()
				cache.On("Lookup", mock.Anything, mock.Anything, false).
					Return(nil, false, nil).Once()

				tags.On("FindKeysByTags", mock.Anything, "tag-a").
					Return([]string{"entry"}, nil).Once()

				repo.On("ReadRequest", mock.Anything, "entry").
					Return(buildDatabaseEntry(
						sampleFullODTItem("1", "package", "FRA", "BER"),
					), nil).Once()

				repo.On("CreateRequest", mock.Anything, mock.Anything).
					Return(nil).Once()

				cache.On(
					"Store",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					true,
					false,
				).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "deterministic validation failure → abort",
			tags: "tag-a",
			body: sampleRequestBody("package", []string{"FRA"}),
			setupMock: func(
				cache *mocks.MockCacheService,
				tags *mocks.MockTagSearchService,
				repo *mockRepo.MockDatabaseRepository,
			) {
				cache.On("Lookup", mock.Anything, mock.Anything, true).
					Return(nil, false, nil).Once()
				cache.On("Lookup", mock.Anything, mock.Anything, false).
					Return(nil, false, nil).Once()

				tags.On("FindKeysByTags", mock.Anything, "tag-a").
					Return([]string{"entry"}, nil).Once()

				repo.On("ReadRequest", mock.Anything, "entry").
					Return(buildDatabaseEntry(
						sampleODTItemWithoutRooms("1", "package", "FRA", "BER"),
					), nil).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := mocks.NewMockCacheService(t)
			tagSvc := mocks.NewMockTagSearchService(t)
			repo := mockRepo.NewMockDatabaseRepository(t)

			if tt.setupMock != nil {
				tt.setupMock(cache, tagSvc, repo)
			}

			svc, err := NewResponseUpdateService(
				logger,
				tracer,
				tagSvc,
				repo,
				cache,
			)
			assert.NoError(t, err)

			req := buildRequest(tt.body, tt.tags)
			err = svc.UpdateResponse(context.Background(), &req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSelectRelevantOffers_EdgeCases does validate guard clauses and edge conditions.
func TestSelectRelevantOffers_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		resp        *entity.ODTResponse
		body        *entity.RequestBody
		wantErr     bool
		expectPanic bool
	}{
		{
			name:        "nil response",
			resp:        nil,
			body:        &entity.RequestBody{},
			wantErr:     true,
			expectPanic: false,
		},
		{
			name: "empty items",
			resp: &entity.ODTResponse{
				Data: struct {
					Items []entity.ODTItem `json:"items"`
				}{Items: nil},
			},
			body:        &entity.RequestBody{},
			wantErr:     true,
			expectPanic: false,
		},
		{
			name: "nil request body causes panic (programming error)",
			resp: &entity.ODTResponse{Data: struct {
				Items []entity.ODTItem `json:"items"`
			}{Items: []entity.ODTItem{sampleItem("1", "package", "FRA")}}},
			body:        nil,
			expectPanic: true,
		},
		{
			name: "empty request body allows all offers",
			resp: func() *entity.ODTResponse {
				response := buildResponse(
					sampleItem("1", "package", "FRA"),
					sampleItem("2", "flight", "MUC"),
				)
				return &response
			}(),
			body:        &entity.RequestBody{},
			wantErr:     false,
			expectPanic: false,
		},
		{
			name: "airport code case insensitive",
			resp: func() *entity.ODTResponse {
				response := buildResponse(
					sampleItem("1", "package", "fra"),
				)
				return &response
			}(),
			body: &entity.RequestBody{
				DepartureAirportList: []string{"FRA"},
			},
			wantErr:     false,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					_, _ = selectRelevantOffers(tt.resp, tt.body)
				})
				return
			}

			items, err := selectRelevantOffers(tt.resp, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, items)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, items)
			}
		})
	}
}

// TestUpdateResponse_InputValidation does verify early request validation failures.
func TestUpdateResponse_InputValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		req     *entity.Request
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty body",
			req: &entity.Request{
				Body: "",
			},
			wantErr: true,
		},
		{
			name: "invalid json body",
			req: &entity.Request{
				Body: "{not-json}",
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			req: &entity.Request{
				Body: `{"departuredate":"2025-01-01"}`,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := NewResponseUpdateService(
				logger,
				tracer,
				mocks.NewMockTagSearchService(t),
				mockRepo.NewMockDatabaseRepository(t),
				mocks.NewMockCacheService(t),
			)

			err := svc.UpdateResponse(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// nolint: funlen
// TestValidateODTItem tests the validation of an item
func TestValidateODTItem(t *testing.T) {
	validItem := sampleFullODTItem("offer-1", "package", "FRA", "BER")
	validItem.DepartureDate = "2025-07-01T00:00:00Z"
	validItem.ReturnDate = "2025-07-08T00:00:00Z"
	validItem.Accommodation.CheckInDate = validItem.DepartureDate
	validItem.Accommodation.CheckOutDate = validItem.ReturnDate
	validItem.OvernightDuration.NightsInHotel = 7
	validItem.TravelType = "package"

	body := &entity.RequestBody{TravelType: "package"}

	tests := []struct {
		name    string
		item    *entity.ODTItem
		mutate  func(*entity.ODTItem)
		wantErr bool
	}{
		{
			name:    "nil item",
			item:    nil,
			wantErr: true,
		},

		// --- Mandatory field checks ---
		{
			name: "missing offer id",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.OfferID = ""
			},
			wantErr: true,
		},
		{
			name: "missing departure date",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.DepartureDate = ""
			},
			wantErr: true,
		},
		{
			name: "missing return date",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.ReturnDate = ""
			},
			wantErr: true,
		},
		{
			name: "missing accommodation check-in",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Accommodation.CheckInDate = ""
			},
			wantErr: true,
		},
		{
			name: "missing accommodation check-out",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Accommodation.CheckOutDate = ""
			},
			wantErr: true,
		},
		{
			name: "empty rooms",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Accommodation.Rooms = nil
			},
			wantErr: true,
		},
		{
			name: "missing outbound airport",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Flight.OutboundDepartureAirport.Code = ""
			},
			wantErr: true,
		},
		{
			name: "missing inbound airport",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Flight.InboundDepartureAirport.Code = ""
			},
			wantErr: true,
		},

		// --- Date & consistency checks ---
		{
			name: "invalid departure date format",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.DepartureDate = "invalid"
			},
			wantErr: true,
		},
		{
			name: "return before departure",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.ReturnDate = "2025-06-01T00:00:00Z"
			},
			wantErr: true,
		},
		{
			name: "checkout before checkin",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.Accommodation.CheckOutDate = "2025-06-30T00:00:00Z"
			},
			wantErr: true,
		},
		{
			name: "overnight duration mismatch",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.OvernightDuration.NightsInHotel = 3
			},
			wantErr: true,
		},
		{
			name: "travel type mismatch",
			item: &validItem,
			mutate: func(i *entity.ODTItem) {
				i.TravelType = "flight"
			},
			wantErr: true,
		},

		// --- Happy path ---
		{
			name:    "all validations pass",
			item:    &validItem,
			mutate:  func(i *entity.ODTItem) {},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item == nil {
				err := validateODTItem(nil, body)
				assert.Error(t, err)
				return
			}

			copy := *tt.item
			if tt.mutate != nil {
				tt.mutate(&copy)
			}

			err := validateODTItem(&copy, body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNormalizeRequestDates does verify request date normalization logic.
func TestNormalizeRequestDates(t *testing.T) {
	tests := []struct {
		name    string
		body    *entity.RequestBody
		wantErr bool
	}{
		{
			name:    "nil body",
			body:    nil,
			wantErr: true,
		},
		{
			name: "invalid format",
			body: &entity.RequestBody{
				DepartureDate: "2025/01/01",
				ReturnDate:    "2025-01-02",
			},
			wantErr: true,
		},
		{
			name: "return before departure",
			body: &entity.RequestBody{
				DepartureDate: "2025-01-02",
				ReturnDate:    "2025-01-01",
			},
			wantErr: true,
		},
		{
			name: "valid dates",
			body: &entity.RequestBody{
				DepartureDate: "2025-01-01",
				ReturnDate:    "2025-01-05",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := normalizeRequestDates(tt.body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helpers
func buildResponse(items ...entity.ODTItem) entity.ODTResponse {
	return entity.ODTResponse{
		Data: struct {
			Items []entity.ODTItem `json:"items"`
		}{
			Items: items,
		},
	}
}

func sampleItem(id, travelType, outbound string) entity.ODTItem {
	return entity.ODTItem{
		OfferID:    id,
		TravelType: travelType,
		Flight: entity.FlightInfo{
			OutboundDepartureAirport: entity.Airport{Code: outbound},
		},
	}
}

func sampleRequestBody(travelType string, airports []string) entity.RequestBody {
	return entity.RequestBody{
		DepartureAirportList: airports,
		DepartureDate:        "2025-07-07",
		ReturnDate:           "2025-07-14",
		TravelType:           travelType,
		Travelers: map[string]entity.Traveler{
			"a1": {Adults: 2},
		},
	}
}

func buildRequest(body entity.RequestBody, tags string) entity.Request {
	raw, _ := json.Marshal(body)
	return entity.Request{
		Header: map[string]string{"content-type": "application/json"},
		Body:   string(raw),
		Tags:   tags,
	}
}

func buildDatabaseEntry(items ...entity.ODTItem) *entity.DatabaseEntry {
	payload := struct {
		HTTPStatusCode int `json:"httpstatuscode"`
		entity.ODTResponse
	}{
		HTTPStatusCode: http.StatusOK,
		ODTResponse: entity.ODTResponse{
			Data: struct {
				Items []entity.ODTItem `json:"items"`
			}{Items: items},
		},
	}

	raw, _ := json.Marshal(payload)
	return &entity.DatabaseEntry{
		Response: entity.Response{Response: string(raw)},
	}
}

func sampleFullODTItem(id, travelType, outbound, inbound string) entity.ODTItem {
	item := sampleItem(id, travelType, outbound)
	item.Flight.InboundDepartureAirport = entity.Airport{Code: inbound}
	item.Accommodation.Rooms = []entity.AccommodationRoom{{ID: 1}}
	item.Description = "ok"
	item.Price = entity.PriceInfo{Amount: 100, Currency: "EUR"}
	return item
}

func sampleODTItemWithoutRooms(id, travelType, outbound, inbound string) entity.ODTItem {
	item := sampleFullODTItem(id, travelType, outbound, inbound)
	item.Accommodation.Rooms = nil
	return item
}
