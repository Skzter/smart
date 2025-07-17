package service

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
)

// Test for Service
func TestNewService(t *testing.T) {
	tests := []struct {
		testName      string
		logger        *slog.Logger
		expectedError bool
	}{
		{
			testName:      "Invalid Logger",
			logger:        nil,
			expectedError: true,
		},
		{
			testName:      "Valid Parameter",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service, err := NewOpenAI(test.logger, repo.NewMockOpenAI(t))

			if test.expectedError {
				if err == nil {
					t.Errorf("WARNING: Expected Error, but nil")
				}

				if service != nil {
					t.Error("WARNING: expected server has to be nil, but received service")
				}
			} else if err != nil {
				t.Errorf("WARNING: expected no error, but received error")
			}
		})
	}
}

// Test for request
func TestRequest(t *testing.T) {
	tests := []struct {
		testName      string
		request       entity.Request
		expectedError bool
	}{
		{
			testName: "Nil-Content",
			request: entity.Request{
				Prompt:       "Test",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: true,
		},
		{
			testName: "Valid Request",
			request: entity.Request{
				Prompt:       "Test",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			mockOpenAiRepo := repo.NewMockOpenAI(t)
			if test.expectedError {
				mockOpenAiRepo.On("CreateRequest", mock.Anything, test.request).Return(nil, fmt.Errorf("Expected Error"))
			} else {
				mockOpenAiRepo.On("CreateRequest", mock.Anything, test.request).Return(&entity.Response{Text: "Test", SessionID: "123 Test"}, nil)
			}

			service, err := NewOpenAI(logger, mockOpenAiRepo)

			if err != nil {
				t.Errorf("WARNING: Failed to create openAIService")
			}

			resp, err := service.Request(t.Context(), test.request)

			if test.expectedError {
				if err == nil {
					t.Errorf("WARNING: Expected Error")
				}
				if resp != nil {
					t.Errorf("WARNING: Expected Error")
				}
			} else {
				if err != nil {
					t.Errorf("WARNING: Unexpected Error")
				}
				if resp == nil {
					t.Errorf("WARNING: Unexpected Error, expected response")
				}
			}
		})
	}
}
