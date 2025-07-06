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

// TestValidatorValidate tests the validation method of the validator
func TestValidatorValidate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Test configuration
	cfg := &config.Config{
		Model: "gpt-4o",
		Prompts: &config.Prompts{
			ValidationPrompt: "test prompt",
		},
		Timeout: 10,
	}

	validator := service.NewValidator(logger, cfg)

	// Test cases as a table
	tests := []struct {
		name            string
		input           string
		expectCall      bool
		expectedContent string
	}{
		{
			name:       "empty JSON string",
			input:      "",
			expectCall: false,
		},
		{
			name:       "non-200 status",
			input:      `{"httpstatuscode":404}`,
			expectCall: false,
		},
		{
			name: "valid 200 response",
			input: func() string {
				vr := entity.ValidationResponse{
					HTTPStatusCode: 200,
					Data: entity.ValidationData{
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
		},
	}

	// Execute tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConnector := mockrepo.NewOpenAI(t)

			if tt.expectCall {
				mockConnector.
					On("CreateRequest", mock.Anything, mock.MatchedBy(func(req entity.Request) bool {
						return strings.Contains(req.Prompt, tt.expectedContent)
					})).
					Return(&entity.Response{Text: "mocked response"}, nil).
					Once()
			}

			validator.Connector = mockConnector

			err := validator.Validate(tt.input)
			if (err != nil) != !tt.expectCall {
				t.Errorf("Validate() error = %v, expectCall %v", err, tt.expectCall)
			}

			mockConnector.AssertExpectations(t)
		})
	}
}
