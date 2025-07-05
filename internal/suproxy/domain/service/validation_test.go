package service_test

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	mockrepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name            string
		input           *service.Response
		expectCall      bool
		expectedContent string
	}{
		{
			name:       "nil response",
			input:      nil,
			expectCall: false,
		},
		{
			name: "non-200 status",
			input: &service.Response{
				HTTPStatusCode: 404,
			},
			expectCall: false,
		},
		{
			name: "valid 200 response",
			input: &service.Response{
				HTTPStatusCode: 200,
				Data: service.Data{
					Items: []service.Item{
						{
							Duration:      5,
							DepartureDate: "2025-07-01",
							ReturnDate:    "2025-07-10",
						},
					},
				},
			},
			expectCall:      true,
			expectedContent: "Duration: 5 | DepartureDate: 2025-07-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConnector := mockrepo.NewOpenAI(t)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			if tt.expectCall {
				mockConnector.
					On("CreateRequest", mock.Anything, mock.MatchedBy(func(req entity.Request) bool {
						return strings.Contains(req.Prompt, tt.expectedContent)
					})).
					Return(&entity.Response{Text: "mocked"}, nil).
					Once()
			}

			validator := service.Validator{
				Connector: mockConnector,
				Logger:    logger,
			}

			validator.Validate(tt.input)

			mockConnector.AssertExpectations(t)
		})
	}
}
