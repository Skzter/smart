package handler

import (
	"encoding/json"
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

//nolint:funlen
func TestHandlerPostOfferlist(t *testing.T) {
	header := map[string]string{
		"Authorization": "Bearer asdfjsafjaölfaöfsal",
	}
	tests := []struct {
		name                      string
		request                   string
		prompt                    string
		useCorrectAdress          bool
		useValidRequestEntity     bool
		simulatedResponse         string // must contain unique testid whenever expectValidateCall is true
		simulatedResponseCode     int
		expectedResponse          string
		expects200                bool
		expectValidateCall        bool
		expectsSupplierCall       bool
		expectedValidationResult  bool
		simulatedValidationResult error
	}{
		{
			name:                  "valid",
			request:               "{\"apimode\":\"live\",\"id\":\"a0950be9-76ad-4fcb-932d-37660d10b1f8\",\"params\":[]}",
			prompt:                "",
			useCorrectAdress:      true,
			useValidRequestEntity: true,
			simulatedResponse: `{
				"httpstatuscode": 200,
				"data": {
					"items":[
						{
							"offerid": 0
						}
					]
				}
			}`,
			simulatedResponseCode: 200,
			expectedResponse: `{
				"httpstatuscode": 200,
				"data": {
					"items":[
						{
							"offerid": 0
						}
					]
				}
			}`,

			expects200:                true,
			expectValidateCall:        true,
			expectsSupplierCall:       true,
			expectedValidationResult:  true,
			simulatedValidationResult: nil,
		},
		{
			name:                     "invalid request body",
			request:                  "invalid body",
			prompt:                   "",
			useValidRequestEntity:    false,
			expectedResponse:         `{"error":"Invalid request body"}`,
			expects200:               false,
			expectValidateCall:       false,
			expectsSupplierCall:      false,
			expectedValidationResult: false,
		},
		{
			name:                     "invalid address",
			request:                  "{\"apimode\":\"live\",\"id\":\"a0950be9-76ad-4fcb-932d-37660d10b1f8\",\"params\":[]}",
			prompt:                   "",
			useValidRequestEntity:    true,
			expectedResponse:         `{"error":"Invalid request body"}`,
			expects200:               false,
			expectValidateCall:       false,
			expectsSupplierCall:      false,
			expectedValidationResult: false,
		},
		{
			name:                     "supplier 400",
			request:                  "{}",
			prompt:                   "",
			useCorrectAdress:         true,
			useValidRequestEntity:    true,
			simulatedResponse:        "",
			simulatedResponseCode:    400,
			expectedResponse:         "",
			expects200:               true,
			expectValidateCall:       false,
			expectsSupplierCall:      true,
			expectedValidationResult: false,
		},
		{
			name:                     "invalid supplier data",
			request:                  "{}",
			prompt:                   "",
			useCorrectAdress:         true,
			useValidRequestEntity:    true,
			simulatedResponse:        "invalid response",
			simulatedResponseCode:    200,
			expectedResponse:         "invalid response",
			expects200:               true,
			expectValidateCall:       false,
			expectsSupplierCall:      true,
			expectedValidationResult: false,
		},
		{
			name:                  "invalid httpstatuscode in supplierdata",
			request:               "{}",
			prompt:                "",
			useCorrectAdress:      true,
			useValidRequestEntity: true,
			simulatedResponse: `{
				"httpstatuscode": "No Number",
				"data": {
					"items":[
						{
							"offerid": 5
						}
					]
				}
			}`,
			simulatedResponseCode: 200,
			expectedResponse: `{
				"httpstatuscode": "No Number",
				"data": {
					"items":[
						{
							"offerid": 5
						}
					]
				}
			}`,
			expects200:               true,
			expectValidateCall:       false,
			expectsSupplierCall:      true,
			expectedValidationResult: false,
		},
		{
			name:                  "invalid data in supplierdata",
			request:               "{}",
			prompt:                "",
			useCorrectAdress:      true,
			useValidRequestEntity: true,
			simulatedResponse: `{
				"httpstatuscode": 200,
				"data": "not a map"
			}`,
			simulatedResponseCode: 200,
			expectedResponse: `{
				"httpstatuscode": 200,
				"data": "not a map"
			}`,
			expects200:               true,
			expectValidateCall:       false,
			expectsSupplierCall:      true,
			expectedValidationResult: false,
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

	for pos, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer.Clear()
			w := httptest.NewRecorder()

			var server *httptest.Server
			if tt.expectsSupplierCall {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.simulatedResponseCode)
					_, err := w.Write([]byte(tt.simulatedResponse))
					if err != nil {
						panic(err)
					}
				}))
				defer server.Close()
			}

			if tt.expectValidateCall {
				mockValidator.On("Validate", mock.Anything, mock.MatchedBy(func(list *entity.SupplierResponse) bool {
					var data map[string]any
					if err := json.Unmarshal(list.Data.Items[0], &data); err != nil {
						panic(err)
					}
					return int(data["offerid"].(float64)) == pos
				})).Return(tt.simulatedValidationResult)
			}
			h.validator = mockValidator

			var reqstring []byte
			if tt.useValidRequestEntity {
				var dest string
				if tt.useCorrectAdress {
					dest = server.URL
				} else {
					dest = ""
				}

				reqEnt := entity.Request{
					Header:      header,
					Prompt:      tt.prompt,
					Destination: dest,
					Request:     tt.request,
				}
				reqstring, _ = json.Marshal(reqEnt)
			} else {
				reqstring = []byte(tt.request)
			}

			req, _ := http.NewRequest("POST", "/api/v1/Offerlist", strings.NewReader(string(reqstring)))
			router.ServeHTTP(w, req)

			bytes, _ := io.ReadAll(w.Body)

			valid := true
			for i := range writer.len() {
				if strings.Contains(string(writer.Read(i)), "ERROR") {
					valid = false
				}
			}

			assert.Equal(t, tt.expectedValidationResult, valid)
			assert.Equal(t, tt.expects200, w.Code == http.StatusOK)
			assert.Equal(t, tt.expectedResponse, string(bytes))
		})
	}
}
