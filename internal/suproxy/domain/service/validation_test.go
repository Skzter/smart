package service_test

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	mockrepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// TestValidatorValidate tests the Validate method of the Validator service
// nolint:funlen
func TestValidatorValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := &config.Config{
		Model: "gpt-4o",
		Prompts: &config.Prompts{
			ValidationPrompt: "test prompt",
		},
		Timeout:               20,
		MaxItemsPerValidation: 10,
	}

	validator := service.NewValidator(logger, cfg)

	tests := []struct {
		name            string
		input           string
		expectCall      bool
		expectedContent string
		mockResponse    string
		expectError     bool
	}{
		{
			name:        "empty JSON string",
			input:       "",
			expectCall:  false,
			expectError: true,
		},
		{
			name:        "non-200 status",
			input:       `{"httpstatuscode":404}`,
			expectCall:  false,
			expectError: true,
		},
		{
			name: "valid 200 response with valid OpenAI result",
			input: func() string {
				vr := entity.SupplierOfferResponse{
					HTTPStatusCode: 200,
					Data: entity.SupplierOfferData{
						Items: []json.RawMessage{
							json.RawMessage(`{
								"duration": 5,
								"departuredate": "2025-07-01",
								"returndate": "2025-07-10"
							}`),
						},
					},
				}
				b, _ := json.Marshal(vr)
				return string(b)
			}(),
			expectCall:      true,
			expectedContent: "duration\":5",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectError:     false,
		},
		{
			name: "valid 200 response with invalid OpenAI result",
			input: func() string {
				vr := entity.SupplierOfferResponse{
					HTTPStatusCode: 200,
					Data: entity.SupplierOfferData{
						Items: []json.RawMessage{
							json.RawMessage(`{
								"duration": 7,
								"departuredate": "2025-08-01",
								"returndate": "2025-08-10"
							}`),
						},
					},
				}

				b, _ := json.Marshal(vr)

				return string(b)
			}(),
			expectCall:      true,
			expectedContent: "duration\":7",
			mockResponse:    `{"valid":false,"reason":["missing_hotelid"]}`,
			expectError:     true,
		},
		{
			name: "valid 200 response with invalid JSON from OpenAI",
			input: func() string {
				vr := entity.SupplierOfferResponse{
					HTTPStatusCode: 200,
					Data: entity.SupplierOfferData{
						Items: []json.RawMessage{
							json.RawMessage(`{
								"duration": 9,
								"departuredate": "2025-09-01",
								"returndate": "2025-09-10"
							}`),
						},
					},
				}

				b, _ := json.Marshal(vr)
				return string(b)
			}(),
			expectCall:      true,
			expectedContent: "duration\":9",
			mockResponse:    `invalid json`,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConnector := mockrepo.NewOpenAI(t)

			if tt.expectCall {
				mockConnector.
					On("CreateRequest", mock.Anything, mock.MatchedBy(func(req entity.Request) bool {
						return strings.Contains(req.Prompt, tt.expectedContent)
					})).
					Return(&entity.Response{Text: tt.mockResponse}, nil).
					Once()
			}

			validator.SetOpenAIService(mockConnector)
			err := validator.Validate(tt.input)

			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}

			mockConnector.AssertExpectations(t)
		})
	}
}
