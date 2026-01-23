package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks/service"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

type slicewriter struct {
	data [][]byte
}

func (s *slicewriter) Write(p []byte) (n int, err error) {
	s.data = append(s.data, p)
	return len(p), nil
}

func (s *slicewriter) Clear() {
	s.data = [][]byte{}
}

func (s *slicewriter) Read(pos int) []byte {
	return s.data[pos]
}

func (s *slicewriter) len() int {
	return len(s.data)
}

func RejectValidator(t testing.TB) service.Validator {
	discardValidator := mocks.NewMockValidator(t)
	discardValidator.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("reject")).Maybe()
	return discardValidator
}

func TestNewSuproxyController(t *testing.T) {
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		log     *slog.Logger
		cfg     *config.Config
		val     service.Validator
		clt     *http.Client
		db      service.DatabaseService
		syncer  service.TaglistSync
		metrics sharedService.MetricsService
		cache   service.CacheService
		err     bool
		updater service.ResponseUpdateService
	}{
		{
			name:    "valid",
			cfg:     &config.Config{},
			log:     slog.Default(),
			val:     RejectValidator(t),
			clt:     &http.Client{},
			db:      mocks.NewMockDatabaseService(t),
			syncer:  mocks.NewMockTaglistSync(t),
			metrics: sharedMocks.NewMockMetricsService(t),
			cache:   mocks.NewMockCacheService(t),
			updater: mocks.NewMockResponseUpdateService(t),
			err:     false,
		},
		{
			name:    "params nil",
			log:     nil,
			cfg:     nil,
			val:     nil,
			clt:     nil,
			db:      nil,
			syncer:  nil,
			metrics: nil,
			cache:   nil,
			err:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, err := handler.NewSuproxyController(tt.log, tt.cfg, tt.val, tt.clt, tt.db, tracer, tt.syncer, tt.metrics, tt.cache, tt.updater)

			assert.Equal(t, tt.err, controller == nil)
			assert.Equal(t, tt.err, err != nil)
		})
	}
}

type supplierSetup struct {
	code     int
	response any // if nil, will use expected response
}

//nolint:funlen
func TestHandlerPostOfferlist(t *testing.T) {
	header := map[string]string{
		"Authorization": "Bearer asdfjsafjaölfaöfsal",
	}

	invalidRequestBody := struct {
		Err string `json:"error"`
	}{
		Err: "invalid request body",
	}

	tests := []struct {
		name                 string
		request              *entity.Request // will use invalid request if nil
		useCorrectAdress     bool
		sSetup               *supplierSetup // only sets up server if not nil
		tsSetup              *[]any
		expectedResponse     any // allows for unmarshal to fail
		expects200           bool
		expectGetTaglistCall bool
	}{
		{
			name: "valid",
			request: &entity.Request{
				Tags: "",
				Body: `{}`,
			},
			useCorrectAdress: true,

			expectedResponse: entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{[]byte(`{"offerid": 213213}`)},
				},
			},
			expects200: true,

			sSetup: &supplierSetup{
				code:     200,
				response: nil,
			},
			expectGetTaglistCall: true,
		},
		{
			name:             "invalid request body",
			request:          nil,
			expectedResponse: invalidRequestBody,
			expects200:       false,
		},
		{
			name: "invalid address",
			request: &entity.Request{
				Tags: "",
				Body: `{}`,
			},
			useCorrectAdress: false,
			expectedResponse: invalidRequestBody,
			expects200:       false,
		},
		{
			name:             "supplier 400",
			request:          &entity.Request{},
			useCorrectAdress: true,
			expectedResponse: entity.SupplierResponse{
				HTTPStatusCode: 0,
			},
			expects200: false,
			sSetup: &supplierSetup{
				code:     400,
				response: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := RejectValidator(t)
			mockDB := mocks.NewMockDatabaseService(t)
			mockSyncer := mocks.NewMockTaglistSync(t)
			mockMetrics := sharedMocks.NewMockMetricsService(t)
			mockCache := mocks.NewMockCacheService(t)
			mockUpdate := mocks.NewMockResponseUpdateService(t)
			tracer := otel.Tracer("test")

			// Setup metrics mock to accept any calls
			mockMetrics.On("IncRequestSuccess").Return().Maybe()
			mockMetrics.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetrics.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetrics.On("RecordStatusCode", mock.Anything).Return().Maybe()
			mockCache.On(
				"Lookup",
				mock.Anything, // context
				mock.Anything, // entity.Request
				mock.Anything, // isMock bool
			).Return([]byte(nil), false, nil).Maybe()
			mockCache.On(
				"Store",
				mock.Anything, // context
				mock.Anything, // entity.Request
				mock.Anything, // response []byte
				mock.Anything, // isMock bool
				mock.Anything, // isError bool
			).Return(nil).Maybe()
			mockSyncer.
				On("GetCurrentTaglist").
				Return(&sharedEntity.TagList{}, nil).
				Maybe()

			h, _ := handler.NewSuproxyController(slog.New(slog.DiscardHandler), &config.Config{}, validator, &http.Client{}, mockDB, tracer, mockSyncer, mockMetrics, mockCache, mockUpdate)

			router := SetupRouter(h)
			w := httptest.NewRecorder()

			var server *httptest.Server
			if tt.sSetup != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.sSetup.code)
					if tt.sSetup.response == nil {
						tt.sSetup.response = tt.expectedResponse
					}
					body, _ := json.Marshal(tt.sSetup.response)
					_, err := w.Write(body)
					if err != nil {
						panic(err)
					}
				}))
				defer server.Close()
			}

			// use correct header and destintation if request is not nil
			var reqstring []byte
			if tt.request != nil {
				var dest string
				if tt.useCorrectAdress {
					dest = server.URL
				} else {
					dest = ""
				}
				tt.request.Header = header
				tt.request.Destination = dest
				reqstring, _ = json.Marshal(tt.request)
			} else {
				reqstring = []byte("invalid")
			}

			req, _ := http.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader(string(reqstring)))
			router.ServeHTTP(w, req)

			bytes, _ := io.ReadAll(w.Body)
			expectedBytes, _ := json.Marshal(tt.expectedResponse)

			assert.Equal(t, tt.expects200, w.Code == http.StatusOK)
			assert.Equal(t, string(expectedBytes), string(bytes))
		})
	}
}

type dbSetup struct {
	err error
}

type validationSetup struct {
	err  error
	tags *sharedEntity.TagList
}

//nolint:funlen
func TestHandlerHandleRequest(t *testing.T) {
	validRespData, err := json.Marshal(entity.SupplierResponse{
		HTTPStatusCode: 200,
		Data: entity.SupplierOfferList{
			Items: []json.RawMessage{
				[]byte(`{"offerid":223213}`),
			},
		},
	})
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name                 string
		respData             []byte
		dbSetup              *dbSetup
		vsetup               *validationSetup
		wantSyncEr           bool
		expectLoggedError    bool
		expectGetTaglistCall bool
	}{
		{
			name:     "valid, sucessful storage",
			respData: validRespData,
			vsetup: &validationSetup{
				err:  nil,
				tags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "valid", Description: ""}}},
			},
			dbSetup:              &dbSetup{err: nil},
			wantSyncEr:           false,
			expectLoggedError:    false,
			expectGetTaglistCall: true,
		},
		{
			name:     "valid, storage error",
			respData: validRespData,
			vsetup: &validationSetup{
				err:  nil,
				tags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "valid", Description: ""}}},
			},
			dbSetup:              &dbSetup{err: errors.New("Storage error")},
			wantSyncEr:           false,
			expectLoggedError:    true,
			expectGetTaglistCall: true,
		},
		{
			name:                 "invalid resp data",
			respData:             []byte("invalid"),
			expectLoggedError:    true,
			expectGetTaglistCall: false,
		},
		{
			name:     "sync error",
			respData: validRespData,
			vsetup: &validationSetup{
				err:  nil,
				tags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "valid", Description: ""}}},
			},
			dbSetup:              &dbSetup{err: nil},
			wantSyncEr:           true,
			expectLoggedError:    true,
			expectGetTaglistCall: true,
		},
	}

	mockValidator := mocks.NewMockValidator(t)
	mockDB := mocks.NewMockDatabaseService(t)
	mockSyncer := mocks.NewMockTaglistSync(t)
	mockMetrics := sharedMocks.NewMockMetricsService(t)
	mockCache := mocks.NewMockCacheService(t)
	mockUpdate := mocks.NewMockResponseUpdateService(t)
	tracer := otel.Tracer("test")

	var writer slicewriter

	h, _ := handler.NewSuproxyController(slog.New(slog.NewJSONHandler(&writer, nil)), &config.Config{}, mockValidator, &http.Client{}, mockDB, tracer, mockSyncer, mockMetrics, mockCache, mockUpdate)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.Clear()

			if tt.vsetup != nil {
				mockValidator.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(tt.vsetup.tags, tt.vsetup.err)
				defer func() {
					mockValidator.AssertExpectations(t)
					mockValidator.ExpectedCalls = []*mock.Call{}
				}()
			}

			if tt.dbSetup != nil {
				mockDB.On("SaveDbEntry", mock.Anything, mock.Anything).Return(tt.dbSetup.err)
				defer func() {
					mockDB.AssertExpectations(t)
					mockDB.ExpectedCalls = []*mock.Call{}
				}()
			}
			if tt.expectGetTaglistCall {
				mockSyncer.On("GetCurrentTaglist").
					Return(&sharedEntity.TagList{
						Tags: []sharedEntity.Tag{
							{Name: "ResponseNot200", Description: "response not 200"},
						},
					}, nil).
					Maybe()
			}
			if tt.wantSyncEr {
				mockSyncer.On("SyncTaglist", mock.Anything, mock.Anything).Return(errors.New("syncing error"))
				defer func() {
					mockSyncer.AssertExpectations(t)
					mockSyncer.ExpectedCalls = []*mock.Call{}
				}()
			} else {
				mockSyncer.On("SyncTaglist", mock.Anything, mock.Anything).Return(nil)
				defer func() {
					mockSyncer.AssertExpectations(t)
					mockSyncer.ExpectedCalls = []*mock.Call{}
				}()
			}

			h.HandleRequest(t.Context(), entity.Request{}, &tt.respData)

			err := false
			for i := 0; i < writer.len(); i++ {
				if strings.Contains(string(writer.Read(i)), "ERROR") {
					err = true
				}
				t.Log("ERROR: ", string(writer.Read(i)))
			}

			assert.Equal(t, tt.expectLoggedError, err)
		})
	}
}

func BenchmarkPostOfferList(b *testing.B) {
	tracer := otel.Tracer("test")

	cache := mocks.NewMockCacheService(b)
	mockUpdater := mocks.NewMockResponseUpdateService(b)
	mockMetrics := sharedMocks.NewMockMetricsService(b)
	mockDB := mocks.NewMockDatabaseService(b)
	mockSyncer := mocks.NewMockTaglistSync(b)

	mockMetrics.On("IncRequestSuccess").Return().Maybe()
	mockMetrics.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetrics.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetrics.On("RecordStatusCode", mock.Anything).Return().Maybe()

	cache.On(
		"Lookup",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return([]byte(nil), false, nil).Maybe()

	cache.On(
		"Store",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(nil).Maybe()

	mockSyncer.
		On("GetCurrentTaglist").
		Return(&sharedEntity.TagList{}, nil).
		Maybe()

	ctrl, _ := handler.NewSuproxyController(
		slog.New(slog.DiscardHandler),
		&config.Config{},
		RejectValidator(b),
		&http.Client{},
		mockDB,
		tracer,
		mockSyncer,
		mockMetrics,
		cache,
		mockUpdater,
	)

	router := SetupRouter(ctrl)

	response := entity.SupplierResponse{
		HTTPStatusCode: 200,
		Data: entity.SupplierOfferList{
			Items: []json.RawMessage{
				[]byte(`{"offerid": 213213}`),
			},
		},
	}
	mResponse, _ := json.Marshal(response)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(mResponse)
	}))
	defer server.Close()

	requestBody, _ := json.Marshal(entity.Request{
		Header:      map[string]string{"Content-Type": "application/json"},
		Tags:        "",
		Destination: server.URL,
		Body:        `{"apimode":"live","id":"a0950be9-76ad-4fcb-932d-37660d10b1f8","params":[]}`,
	})

	b.ResetTimer()

	for b.Loop() {
		req, _ := http.NewRequest(
			http.MethodPost,
			"/api/v1/Offerlist",
			bytes.NewReader(requestBody),
		)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status code: got %d", w.Code)
		}
	}
}

// setupRouter initializes the Gin router and sets up the routes for the API
func SetupRouter(h *handler.SuproxyController) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", h.PostOfferlist)
	}

	return router
}
