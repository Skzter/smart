package service

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	mockrepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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

	validator := NewValidator(logger, cfg)

	tests := []struct {
		name            string
		input           *entity.SupplierOfferList
		expectCall      bool
		expectedContent string
		mockResponse    string
		expectError     bool
	}{
		{
			name:        "empty JSON string",
			input:       nil,
			expectCall:  false,
			expectError: true,
		},
		{
			name: "non-200 status",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Offers:         []entity.SupplierOffer{},
			},
			expectCall:  false,
			expectError: true,
		},
		{
			name: "valid 200 response with valid OpenAI result",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Offers: []entity.SupplierOffer{{Data: []byte(`{
					"duration": 5,
					"departuredate": "2025-07-01",
					"returndate": "2025-07-10",
					}`)},
				},
			},
			expectCall:      true,
			expectedContent: "duration\":5",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectError:     false,
		},
		{
			name: "valid 200 response with invalid OpenAI result",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Offers: []entity.SupplierOffer{{Data: []byte(`{
						"duration": 7,
						"departuredate": "2025-08-01",
						"returndate": "2025-08-10"
						}`)},
				},
			},
			expectCall:      true,
			expectedContent: "duration\":7",
			mockResponse:    `{"valid":false,"reason":["missing_hotelid"]}`,
			expectError:     true,
		},
		{
			name: "valid 200 response with invalid JSON from OpenAI",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Offers: []entity.SupplierOffer{{Data: []byte(`{
						"duration": 9,
						"departuredate": "2025-09-01",
						"returndate": "2025-09-10"
					}`)},
				},
			},
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
					Return(&entity.Response{Response: tt.mockResponse}, nil).
					Once()
			}

			validator.SetOpenAIService(mockConnector)
			err := validator.Validate(t.Context(), tt.input)

			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}

			mockConnector.AssertExpectations(t)
		})
	}
}
