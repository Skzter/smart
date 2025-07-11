package service

/*
import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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

	tests := []struct {
		name        string
		cfg         *config.Config
		logger      *slog.Logger
		expectError bool
	}{
		{
			name:        "valid",
			cfg:         &cfg,
			logger:      logger,
			expectError: false,
		},
		{
			name:        "nil config",
			cfg:         nil,
			logger:      logger,
			expectError: true,
		},
		{
			name:        "nil logger",
			cfg:         &cfg,
			logger:      nil,
			expectError: true,
		},
		{
			name: "NewService error",
			cfg: &config.Config{
				Timeout: 0,
				Prompts: &config.Prompts{
					ValidationPrompt: "validate this",
				},
				Model:                 "gpt-4",
				MaxItemsPerValidation: 5,
			},
			logger:      logger,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewValidator(tt.logger, tt.cfg)
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
		input            *entity.SupplierOfferList
		expectCall       bool
		expectedContent  string
		mockResponse     string
		mockResonseError error
		expectError      bool
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
				Data:           &map[string]any{},
			},
			expectCall:  false,
			expectError: true,
		},
		{
			name: "valid 200 response with valid OpenAI result",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"duration":      5,
							"departuredate": "2025-01-01",
							"returndate":    "2025-01-10",
						},
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-01-01",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectError:     false,
		},
		{
			name: "valid 200 response with invalid OpenAI result",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"duration":      7,
							"departuredate": "2025-02-01",
							"returndate":    "2025-02-10",
						},
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-02-01",
			mockResponse:    `{"valid":false,"reason":["missing_hotelid"]}`,
			expectError:     true,
		},
		{
			name: "valid 200 response with invalid JSON from OpenAI",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"duration":      9,
							"departuredate": "2025-03-01",
							"returndate":    "2025-03-10",
						},
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-03-01",
			mockResponse:    `invalid json`,
			expectError:     true,
		},
		{
			name: "valid 400 response",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 400,
				Data:           &map[string]any{},
			},
			expectCall:  false,
			expectError: true,
		},
		{
			name: "valid 200 response with empty offers",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{},
				},
			},
			expectCall:  false,
			expectError: true,
		},
		{
			name: "exceeding maximum request,",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"departuredate": "2025-05-01",
						},
						map[string]any{
							"departuredate": "2025-05-01",
						},
						map[string]any{
							"departuredate": "2025-05-01",
						},
						map[string]any{
							"departuredate": "2025-05-01",
						},
						map[string]any{
							"departuredate": "2025-05-01",
						},
						map[string]any{
							"departuredate": "2025-05-01",
						},
					},
				},
			},
			expectCall:      true,
			expectedContent: "2025-05-01",
			mockResponse:    `{"valid":true,"reason":[]}`,
			expectError:     false,
		},
		{
			name: "valid 200 response with single empty offer",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{},
					},
				},
			},
			expectCall:  false,
			expectError: false,
		},
		{
			name: "openai service error",
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"departuredate": "2025-06-01",
						},
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
			input: &entity.SupplierOfferList{
				HTTPStatusCode: 200,
				Data: &map[string]any{
					"items": []any{
						map[string]any{
							"departuredate": "2025-07-01",
						},
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockservice := mocks.NewMockOpenAIService(t)

			if tt.expectCall {
				mockservice.
					On("Request", mock.Anything, mock.MatchedBy(func(req shared.Request) bool {
						return strings.Contains(req.Prompt, tt.expectedContent)
					})).
					Return(&shared.Response{Text: tt.mockResponse}, tt.mockResonseError)
			}

			validator := Validator{
				mockservice,
				logger,
				"",
				"",
				5,
			}

			err := validator.Validate(t.Context(), tt.input)

			if (err != nil) != tt.expectError {
				t.Errorf("Validate() error = %v, expectError %v", err, tt.expectError)
			}

			mockservice.AssertExpectations(t)
		})
	}
}*/
