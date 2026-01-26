package tools

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/service"
)

//nolint:dupl,funlen
func TestNewRunTestTool(t *testing.T) {
	tests := []struct {
		name              string
		logger            *slog.Logger
		autotesterService *mocks.MockAutotesterAPIService
		expectedError     bool
		expectedErrorMsg  string
	}{
		{
			name:              "successful creation",
			logger:            slog.New(slog.DiscardHandler),
			autotesterService: mocks.NewMockAutotesterAPIService(t),
			expectedError:     false,
		},
		{
			name:              "nil logger",
			logger:            nil,
			autotesterService: mocks.NewMockAutotesterAPIService(t),
			expectedError:     true,
			expectedErrorMsg:  "nil",
		},
		{
			name:              "nil autotester service",
			logger:            slog.New(slog.DiscardHandler),
			autotesterService: nil,
			expectedError:     true,
			expectedErrorMsg:  "nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool, err := NewRunTestTool(test.logger, test.autotesterService)

			if test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, tool)
				if test.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), test.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tool)
				assert.Equal(t, test.logger, tool.logger)
				assert.Equal(t, test.autotesterService, tool.autotesterAPIService)
			}
		})
	}
}

//nolint:dupl,funlen
func TestRunTestTool_RunTest(t *testing.T) {
	tests := []struct {
		name           string
		input          entity.ExecuteTestRequest
		mockSetup      func(*mocks.MockAutotesterAPIService)
		expectedError  bool
		expectedOutput entity.RunTestResponse
	}{
		{
			name: "successful test execution - passed",
			input: entity.ExecuteTestRequest{
				UserId: "user123",
				ChatId: "chat456",
				Test:   "describe('Login', () => { it('should login', () => { expect(true).toBe(true); }); });",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.ExecuteTestResponse{
					Result: "PASSED",
				}
				m.EXPECT().ExecuteTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.RunTestResponse{
				Result: "PASSED",
			},
		},
		{
			name: "successful test execution - failed",
			input: entity.ExecuteTestRequest{
				UserId: "user789",
				ChatId: "chat999",
				Test:   "describe('Broken Test', () => { it('should fail', () => { expect(true).toBe(false); }); });",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.ExecuteTestResponse{
					Result: "FAILED",
				}
				m.EXPECT().ExecuteTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.RunTestResponse{
				Result: "FAILED",
			},
		},
		{
			name: "service returns error",
			input: entity.ExecuteTestRequest{
				UserId: "user000",
				ChatId: "chat000",
				Test:   "invalid test code",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				m.EXPECT().ExecuteTest(mock.Anything, mock.Anything).
					Return(nil, errors.New("test execution failed")).Once()
			},
			expectedError:  true,
			expectedOutput: entity.RunTestResponse{},
		},
		{
			name: "empty test string",
			input: entity.ExecuteTestRequest{
				UserId: "user111",
				ChatId: "chat222",
				Test:   "",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.ExecuteTestResponse{
					Result: "ERROR: empty test",
				}
				m.EXPECT().ExecuteTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.RunTestResponse{
				Result: "ERROR: empty test",
			},
		},
		{
			name: "test execution with detailed result",
			input: entity.ExecuteTestRequest{
				UserId: "user555",
				ChatId: "chat666",
				Test:   "describe('Complex Test', () => { it('should work', () => { /* complex logic */ }); });",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.ExecuteTestResponse{
					Result: "PASSED: 5 tests passed, 0 failed, duration: 2.5s",
				}
				m.EXPECT().ExecuteTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.RunTestResponse{
				Result: "PASSED: 5 tests passed, 0 failed, duration: 2.5s",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := mocks.NewMockAutotesterAPIService(t)
			test.mockSetup(mockService)

			logger := slog.New(slog.DiscardHandler)
			tool, err := NewRunTestTool(logger, mockService)
			assert.NoError(t, err)

			ctx := context.Background()
			_, output, err := tool.RunTest(ctx, nil, test.input)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput.Result, output.Result)
			}
		})
	}
}
