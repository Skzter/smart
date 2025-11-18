package service_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

func TestNewValidator(t *testing.T) {
	cfg := config.Config{
		Timeout: 10,
		Prompts: &config.Prompts{
			ValidationPrompt: "validate this",
		},
		Model:                 "gpt-4",
		MaxItemsPerValidation: 5,
	}
	logger := slog.Default()
	serv := sharedMocks.NewMockOpenAI(t)

	tests := []struct {
		name        string
		cfg         *config.Config
		logger      *slog.Logger
		service     sharedService.OpenAI
		expectError bool
	}{
		{
			name:        "valid",
			cfg:         &cfg,
			logger:      logger,
			service:     serv,
			expectError: false,
		},
		{
			name:        "nil values",
			cfg:         nil,
			logger:      logger,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := service.NewValidator(tt.logger, tt.cfg, tt.service)
			if !tt.expectError {
				if validator == nil {
					t.Errorf("expected validator, got nil")
				}
				if err != nil {
					t.Error(err)
				}
			} else if validator != nil || err == nil {
				t.Errorf("expected error but got non")
			}
		})
	}
}

// TestValidatorValidate tests the Validate method of the Validator service
// nolint:funlen
func TestValidatorValidate(t *testing.T) {
	tests := []struct {
		name             string
		input            *entity.SupplierResponse
		expectCall       bool
		expectedContent  string
		mockResponse     string
		mockResonseError error
		expectError      bool
		expectedTags     *sharedEntity.TagList
	}{
		{
			name:        "empty input",
			input:       nil,
			expectCall:  false,
			expectError: true,
		},
		{
			name: "non-200 status",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 400,
				Data:           entity.SupplierOfferList{},
			},
			expectCall:   false,
			expectedTags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "non_200", Description: ""}}},
		},
		{
			name: "valid 200 response with valid OpenAI result",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"duration": 7,
						"departuredate": "2025-01-01",
						"returndate":    "2025-01-10"}`),
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-01-01",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectedTags:    &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "valid", Description: ""}}},
		},
		{
			name: "valid 200 response with invalid OpenAI result",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"duration": 7,
						"departuredate": "2025-02-01",
						"returndate":    "2025-02-10"}`),
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-02-01",
			mockResponse:    `{"valid":false,"reason":[ "name": "no_hotelid", "description": "The response does not contain a hotelid field."]}`,
			expectedTags:    &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "no_hotelid", Description: "The response does not contain a hotelid field."}}},
			expectError:     true,
		},
		{
			name: "valid 200 response with invalid JSON from OpenAI",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"duration": 7,
						"departuredate": "2025-03-01",
						"returndate":    "2025-03-10"}`),
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-03-01",
			mockResponse:    `invalid json`,
			expectError:     true,
		},
		{
			name: "valid 200 response with empty offers",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{},
				},
			},
			expectCall:   false,
			expectedTags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "no_offer", Description: ""}}},
		},
		{
			name: "exceeding maximum request,",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
						json.RawMessage(`{"departuredate": "2025-05-01"}`),
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-05-01",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectedTags:    &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "valid", Description: ""}}},
		},
		{
			name: "valid 200 response with single empty offer",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(""),
					},
				},
			},
			expectCall:   false,
			expectedTags: &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "empty_offer", Description: ""}}},
		},
		{
			name: "openai service error",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"duration": 7,
						"departuredate": "2025-06-01",
						"returndate":    "2025-06-10"}`),
					},
				},
			},
			expectCall:       true,
			expectedContent:  "2025-06-01",
			mockResponse:     "",
			expectError:      true,
			mockResonseError: errors.New("error"),
		},
		{
			name: "empty openai response",
			input: &entity.SupplierResponse{
				HTTPStatusCode: 200,
				Data: entity.SupplierOfferList{
					Items: []json.RawMessage{
						json.RawMessage(`{"duration": 7,
						"departuredate": "2025-07-01",
						"returndate":    "2025-07-10"}`),
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-07-01",
			mockResponse:    "",
			expectError:     true,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	cfg := config.Config{
		Timeout:               5,
		MaxItemsPerValidation: 5,
		Prompts: &config.Prompts{
			ValidationPrompt: "validate something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockservice := sharedMocks.NewMockOpenAI(t)

			if tt.expectCall {
				mockservice.
					On("Request", mock.Anything, mock.Anything).
					Return(&sharedEntity.Response{Text: tt.mockResponse}, tt.mockResonseError)
			}

			validator, err := service.NewValidator(logger, &cfg, mockservice)
			if err != nil {
				panic(err)
			}

			tags, err := validator.Validate(t.Context(), tt.input, &sharedEntity.TagList{})

			assert.Equal(t, tt.expectError, err != nil)
			if err != nil {
				assert.Nil(t, tags)
			} else {
				assert.NotNil(t, tags)
				assert.Equal(t, tt.expectedTags, tags)
			}

			mockservice.AssertExpectations(t)
		})
	}
}
