package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// Test for Service
func TestService(t *testing.T) {
	tests := []struct {
		testName      string
		logger        *slog.Logger
		timeout       int
		expectedError bool
	}{
		{
			testName:      "Invalid Logger",
			logger:        nil,
			timeout:       5,
			expectedError: true,
		},
		{

			testName:      "Invalid Timeout for Repository",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			timeout:       0,
			expectedError: true,
		},
		{
			testName:      "Valid Parameter",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			timeout:       5,
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service, err := NewService(test.logger, test.timeout)

			if test.expectedError {
				if err == nil {
					t.Errorf("WARNING: Expected Error, but nil")
				}

				if service != nil {
					t.Error("WARNING: expected server has to be nil, but received service")
				}
			} else if err != nil {
				t.Errorf("WARNING: unexpected Error")
			}
		})
	}
}

// Test for request
func TestRequest(t *testing.T) {
	tests := []struct {
		testName      string
		content       context.Context
		request       entity.Request
		expectedError bool
	}{
		{
			testName: "Nil-logger",
			content:  nil,
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
			content:  context.Background(),
			request: entity.Request{
				Prompt:       "Test",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
	}

	// MockOpenAI := mocks.NewMockOpenAI(t)

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			service, _ := NewService(logger, 5)

			resp, _ := service.Request(test.content, test.request)

			if test.expectedError {
				if resp != nil {
					t.Errorf("WARNING: Expected Error, but received request")
				}
			} /*else if err != nil {
				t.Errorf("WARNING: Unexpected Error")
			}*/
		})
	}
}
