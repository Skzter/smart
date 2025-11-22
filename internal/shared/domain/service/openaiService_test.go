package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
)

// Test for Service
func TestNewService(t *testing.T) {
	tests := []struct {
		testName      string
		repo          repository.OpenAI
		expectedError bool
	}{
		{
			testName:      "nil repo",
			repo:          nil,
			expectedError: true,
		},
		{
			testName:      "Valid Parameter",
			repo:          mocks.NewMockOpenAI(t),
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			service, err := NewOpenAI(test.repo)

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
		testName       string
		requestReturns []any
		request        entity.Request
		expectedError  bool
		ctx            context.Context
	}{
		{
			testName:       "Nil-Context",
			requestReturns: nil,
			expectedError:  true,
			ctx:            nil,
		},
		{
			testName: "Valid Request",
			request: entity.Request{
				Messages: []entity.Message{
					{Role: "user", Body: "user prompt"},
				},
			},
			requestReturns: []any{
				&entity.Message{Role: "assistant", Body: "response"},
				nil,
			},
			expectedError: false,
			ctx:           context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			mockOpenAiRepo := mocks.NewMockOpenAI(t)
			if test.requestReturns != nil {
				mockOpenAiRepo.On("CreateRequest", mock.Anything, mock.Anything).Return(test.requestReturns...)
			}

			service, err := NewOpenAI(mockOpenAiRepo)

			if err != nil {
				t.Errorf("WARNING: Failed to create openAIService")
			}

			resp, err := service.Request(test.ctx, test.request)

			if test.expectedError {
				assert.Nil(t, resp)
				assert.NotNil(t, err)
			} else {
				assert.NotNil(t, resp)
				assert.Nil(t, err)
			}
		})
	}
}
