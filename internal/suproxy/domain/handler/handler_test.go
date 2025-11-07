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
	"time"

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
	discardValidator.On("Validate", mock.Anything, mock.Anything).Return(nil, errors.New("reject")).Maybe()
	return discardValidator
}

func TestNewSuproxyController(t *testing.T) {
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
			controller, err := handler.NewSuproxyController(tt.log, tt.cfg, tt.val, tt.clt, tt.db)

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
	}{
		{
			name: "valid",
			request: &entity.Request{
				//				Prompt:  "",
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
				//				Prompt:  "",
				Request: `{}`,
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

	mockValidator := mocks.NewMockValidator(t)
	mockDB := mocks.NewMockDatabaseService(t)

	h, _ := handler.NewSuproxyController(slog.New(slog.DiscardHandler), &config.Config{}, mockValidator, &http.Client{}, mockDB)

	router := SetupRouter(h)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})

			// Setup mock expectations with completion signal
			if tt.expects200 && tt.sSetup != nil && tt.sSetup.code == 200 {
				// Validator is called first in the async handler
				mockValidator.On("Validate", mock.Anything, mock.Anything).
					Return([]string{"tag1"}, nil).
					Once()

				// Signal completion after last call (to mockDB)
				mockDB.On("SaveDbEntry", mock.Anything, mock.Anything).
					Return(nil).
					Run(func(args mock.Arguments) {
						close(done)
					}).
					Once()
			}

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

			// Wait for async handler if needed
			if tt.expects200 && tt.sSetup != nil && tt.sSetup.code == 200 {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("Timeout waiting for async handler to complete")
				}
			}

			bytes, _ := io.ReadAll(w.Body)
			expectedBytes, _ := json.Marshal(tt.expectedResponse)

			assert.Equal(t, tt.expects200, w.Code == http.StatusOK)
			assert.Equal(t, string(expectedBytes), string(bytes))

			// Clean up
			if tt.expects200 && tt.sSetup != nil && tt.sSetup.code == 200 {
				mockValidator.AssertExpectations(t)
				mockDB.AssertExpectations(t)
				mockValidator.ExpectedCalls = []*mock.Call{}
				mockDB.ExpectedCalls = []*mock.Call{}
			}
		})
	}
}

type dbSetup struct {
	err error
}

type validationSetup struct {
	err  error
	tags []string
}

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
		name              string
		respData          []byte
		dbSetup           *dbSetup
		vsetup            *validationSetup
		expectLoggedError bool
	}{
		{
			name:     "valid, sucessful storage",
			respData: validRespData,
			vsetup: &validationSetup{
				err:  nil,
				tags: []string{"valid"},
			},
			dbSetup:           &dbSetup{err: nil},
			expectLoggedError: false,
		},
		{
			name:     "valid, storage error",
			respData: validRespData,
			vsetup: &validationSetup{
				err:  nil,
				tags: []string{"valid"},
			},
			dbSetup:           &dbSetup{err: errors.New("Storage error")},
			expectLoggedError: true,
		},
		{
			name:              "invalid resp data",
			respData:          []byte("invalid"),
			expectLoggedError: true,
		},
	}

	mockValidator := mocks.NewMockValidator(t)
	mockDB := mocks.NewMockDatabaseService(t)
	var writer slicewriter

	h, _ := handler.NewSuproxyController(slog.New(slog.NewJSONHandler(&writer, nil)), &config.Config{}, mockValidator, &http.Client{}, mockDB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.Clear()

			if tt.vsetup != nil {
				mockValidator.On("Validate", mock.Anything, mock.Anything).Return(tt.vsetup.tags, tt.vsetup.err)
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

			h.HandleRequest(t.Context(), entity.Request{}, &tt.respData)

			err := false
			for i := range writer.len() {
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
	ctrl, _ := handler.NewSuproxyController(
		slog.New(slog.DiscardHandler),
		&config.Config{},
		RejectValidator(b),
		&http.Client{},
		mocks.NewMockDatabaseService(b),
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

	requestBody, _ := json.Marshal(entity.Request{
		Header:      map[string]string{"Content-Type": "application/json"},
		Prompt:      "",
		Destination: server.URL,
		Request:     `{"apimode":"live","id":"a0950be9-76ad-4fcb-932d-37660d10b1f8","params":[]}`,
	})

	for b.Loop() {
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
