package handler

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks"
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

func TestNewSuproxyController(t *testing.T) {
	tests := []struct {
		name   string
		logger *slog.Logger
		cfg    *config.Config
		err    bool
	}{
		{
			name:   "valid",
			logger: slog.Default(),
			cfg: &config.Config{
				Model:                 "gpt",
				Timeout:               10,
				MaxItemsPerValidation: 15,
				Prompts: &config.Prompts{
					ValidationPrompt: "validate this",
				},
			},
			err: false,
		},
		{
			name:   "config nil",
			logger: slog.Default(),
			cfg:    nil,
			err:    true,
		},
		{
			name:   "logger nil",
			logger: nil,
			cfg: &config.Config{
				Model:                 "gpt",
				Timeout:               10,
				MaxItemsPerValidation: 15,
				Prompts: &config.Prompts{
					ValidationPrompt: "validate this",
				},
			},
			err: true,
		},
		{
			name:   "validator error",
			logger: slog.Default(),
			cfg: &config.Config{
				Model:                 "gpt",
				Timeout:               0,
				MaxItemsPerValidation: 15,
				Prompts: &config.Prompts{
					ValidationPrompt: "validate this",
				},
			},
			err: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, err := NewSuproxyController(tt.logger, tt.cfg)

			assert.Equal(t, tt.err, controller == nil)
			assert.Equal(t, tt.err, err != nil)
		})
	}
}

type supplierSetup struct {
	code     int
	response any // if nil, will use expected response
}

type validationSetup struct {
	simulatedResult error
	expectedResult  bool
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
		vSetup           *validationSetup // only sets up validation mock if not nil, otherwise also assumes that error is logged before
		sSetup           *supplierSetup   // only sets up server if not nil
		expectedResponse any              // allows for unmarshal to fail
		expects200       bool
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
			vSetup: &validationSetup{
				simulatedResult: nil,
				expectedResult:  true,
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
				Prompt:  "",
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
			vSetup: &validationSetup{
				simulatedResult: errors.New("error"),
				expectedResult:  false,
			},
		},
	}

	mockValidator := mocks.NewMockValidator(t)
	var writer slicewriter

	h := SuproxyController{
		logger:    slog.New(slog.NewJSONHandler(&writer, nil)),
		config:    &config.Config{},
		client:    &http.Client{},
		validator: mockValidator,
	}
	router := SetupRouter(&h)

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

			if tt.vSetup != nil {
				mockValidator.On("Validate", mock.Anything, mock.Anything).Return(tt.vSetup.simulatedResult)
			}
			h.validator = mockValidator
			defer func() { mockValidator.ExpectedCalls = []*mock.Call{} }()

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

			valid := true
			for i := range writer.len() {
				if strings.Contains(string(writer.Read(i)), "ERROR") {
					valid = false
				}
				t.Log(string(writer.Read(i)))
			}

			assert.Equal(t, tt.vSetup != nil && tt.vSetup.expectedResult, valid)
			assert.Equal(t, tt.expects200, w.Code == http.StatusOK)
			assert.Equal(t, string(expectedBytes), string(bytes))
		})
	}
}

func BenchmarkPostOfferList(b *testing.B) {
	discardValidator := mocks.NewMockValidator(b)
	discardValidator.On("Validate", mock.Anything, mock.Anything).Return(nil)

	ctrl := SuproxyController{
		logger:    slog.New(slog.DiscardHandler),
		config:    &config.Config{},
		client:    &http.Client{},
		validator: discardValidator,
	}
	router := SetupRouter(&ctrl)

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
