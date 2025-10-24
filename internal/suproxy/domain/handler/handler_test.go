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

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service/mocks"
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
	// Validator.Validate returns ([]string, error), so the mock must
	// return a slice (or nil) and an error. Return nil slice and an
	// error to simulate rejection.
	discardValidator.On("Validate", mock.Anything, mock.AnythingOfType("*entity.SupplierResponse")).Return(nil, errors.New("reject")).Maybe()
	return discardValidator
}

func TestNewSuproxyController(t *testing.T) {
	tagSearch := mocks.NewTagSearchService(t)
	tests := []struct {
		name string
		log  *slog.Logger
		cfg  *config.Config
		val  service.Validator
		clt  *http.Client
		db   service.DatabaseService
		err  bool
	}{
		{
			name: "valid",
			cfg:  &config.Config{},
			log:  slog.Default(),
			val:  RejectValidator(t),
			clt:  &http.Client{},
			db:   mocks.NewMockDatabaseService(t),
			err:  false,
		},
		{
			name: "params nil",
			log:  nil,
			cfg:  nil,
			val:  nil,
			clt:  nil,
			db:   nil,
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, err := handler.NewSuproxyController(tt.log, tt.cfg, tt.val, tt.clt, tt.db, tagSearch)

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
		name             string
		request          *entity.Request // will use invalid request if nil
		useCorrectAdress bool
		sSetup           *supplierSetup // only sets up server if not nil
		expectedResponse any            // allows for unmarshal to fail
		expects200       bool
		dbError          *struct{ bool }
		validationError  *struct{ bool }
		loggedError      bool
	}{
		{
			name: "valid",
			request: &entity.Request{
				Prompt:  "",
				Request: `{}`,
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
			validationError: &struct{ bool }{false},
			dbError:         &struct{ bool }{false},
			loggedError:     false,
		},
		{
			name: "valid with malformed list entry",
			request: &entity.Request{
				Prompt:  "",
				Request: `{}`,
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

			validationError: &struct{ bool }{false},
			dbError:         &struct{ bool }{true},
			loggedError:     true,
		},
		{
			name: "valid with storage failure",
			request: &entity.Request{
				Prompt:  "",
				Request: `{}`,
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

			validationError: &struct{ bool }{false},
			dbError:         &struct{ bool }{true},
			loggedError:     true,
		},
		{
			name:             "invalid request body",
			request:          nil,
			expectedResponse: invalidRequestBody,
			expects200:       false,
			loggedError:      true,
		},
		{
			name: "invalid address",
			request: &entity.Request{
				Prompt:  "",
				Request: `{}`,
			},
			useCorrectAdress: false,
			expectedResponse: invalidRequestBody,
			expects200:       false,
			loggedError:      true,
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
			loggedError: true,
		},
		{
			name: "invalid supplier data",
			request: &entity.Request{
				Prompt:  "",
				Request: "{}",
			},
			useCorrectAdress: true,
			expectedResponse: "invalid",
			expects200:       true,
			sSetup: &supplierSetup{
				code:     200,
				response: nil,
			},
			loggedError: true,
		},
		{
			name: "validation error",
			request: &entity.Request{
				Prompt:  "",
				Request: `{}`,
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
			validationError: &struct{ bool }{true},
			loggedError:     true,
		},
	}

	mockValidator := mocks.NewMockValidator(t)
	mockDB := mocks.NewMockDatabaseService(t)
	mockTagSearch := mocks.NewTagSearchService(t)

	// Default mock behavior: when tests don't configure the validator or DB
	// explicitly, return success (no tags, no error) so calls are allowed.
	mockValidator.On("Validate", mock.Anything, mock.AnythingOfType("*entity.SupplierResponse")).Return([]string{}, nil)
	mockDB.On("SaveDbEntry", mock.Anything, mock.Anything).Return(nil)
	var writer slicewriter

	h, _ := handler.NewSuproxyController(slog.New(slog.NewJSONHandler(&writer, nil)), &config.Config{}, mockValidator, &http.Client{}, mockDB, mockTagSearch)

	h.SetHandleAsync(false)
	router := SetupRouter(h)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.Clear()
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

			if tt.validationError != nil {
				err := errors.New("error")
				if !tt.validationError.bool {
					err = nil
				}
				// Validator.Validate returns ([]string, error).
				// When simulating an error, return nil slice + error;
				// otherwise return empty slice + nil.
				if err != nil {
					mockValidator.On("Validate", mock.Anything, mock.AnythingOfType("*entity.SupplierResponse")).Return(nil, err)
				} else {
					mockValidator.On("Validate", mock.Anything, mock.AnythingOfType("*entity.SupplierResponse")).Return([]string{}, nil)
				}
				defer func() { mockValidator.ExpectedCalls = []*mock.Call{} }()
			}

			if tt.dbError != nil {
				err := errors.New("error")
				if !tt.dbError.bool {
					err = nil
				}
				mockDB.On("SaveDbEntry", mock.Anything, mock.Anything).Return(err)
				defer func() { mockDB.ExpectedCalls = []*mock.Call{} }()
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

				// Only set expectation if prompt is not empty
				if strings.TrimSpace(tt.request.Prompt) != "" {
					mockTagSearch.On("FindKeysByTag", mock.Anything, tt.request.Prompt).
						Return([]string{"file1.parquet", "file2.parquet"}, nil).Once()
				}
			} else {
				reqstring = []byte("invalid")
			}

			req, _ := http.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader(string(reqstring)))
			router.ServeHTTP(w, req)

			bytes, _ := io.ReadAll(w.Body)
			expectedBytes, _ := json.Marshal(tt.expectedResponse)

			valid := true
			for i := range writer.len() {
				if strings.Contains(string(writer.Read(i)), "ERROR") {
					valid = false
				}
				t.Log(string(writer.Read(i)))
			}

			assert.Equal(t, tt.loggedError, !valid)
			assert.Equal(t, tt.expects200, w.Code == http.StatusOK)
			assert.Equal(t, string(expectedBytes), string(bytes))

			mockValidator.AssertExpectations(t)
			mockTagSearch.AssertExpectations(t)
		})
	}
}

func BenchmarkPostOfferList(b *testing.B) {
	// Setup Mocks
	mockValidator := RejectValidator(b)
	mockDB := mocks.NewMockDatabaseService(b)
	mockTagSearch := mocks.NewTagSearchService(b)

	mockTagSearch.
		On("FindKeysByTag", mock.Anything, mock.AnythingOfType("string")).
		Return([]string{"file1.parquet", "file2.parquet"}, nil)

	ctrl, _ := handler.NewSuproxyController(
		slog.New(slog.DiscardHandler),
		&config.Config{},
		mockValidator,
		&http.Client{},
		mockDB,
		mockTagSearch,
	)

	router := SetupRouter(ctrl)

	response := entity.SupplierResponse{
		HTTPStatusCode: 200,
		Data: entity.SupplierOfferList{
			Items: []json.RawMessage{[]byte(`{"offerid": 213213}`)},
		},
	}
	mResponse, _ := json.Marshal(response)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(mResponse); err != nil {
			panic(err)
		}
	}))
	defer server.Close()

	requestBody, _ := json.Marshal(entity.Request{
		Header:      map[string]string{"Content-Type": "application/json"},
		Prompt:      "",
		Destination: server.URL,
		Request:     `{"apimode":"live","id":"a0950be9-76ad-4fcb-932d-37660d10b1f8","params":[]}`,
	})

	// Benchmark Loop
	for b.ResetTimer(); b.N > 0; b.N-- {
		request, _ := http.NewRequest(http.MethodPost, "/api/v1/Offerlist", bytes.NewReader(requestBody))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, request)

		if w.Code != http.StatusOK {
			b.Fatalf("Unexpected status code: got %d", w.Code)
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
