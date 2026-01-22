package tools

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/service"
)

//nolint:dupl,funlen
func TestNewGenerateTestTool(t *testing.T) {
	tests := []struct {
		name              string
		logger            *slog.Logger
		autotesterService *mocks.MockAutotesterAPIService
		expectedError     bool
		expectedErrorMsg  string
	}{
		{
			name:              "successful creation",
			logger:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
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
			logger:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
			autotesterService: nil,
			expectedError:     true,
			expectedErrorMsg:  "nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool, err := NewGenerateTestTool(test.logger, test.autotesterService)

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
func TestGenerateTestTool_GenerateTest(t *testing.T) {
	tests := []struct {
		name           string
		input          entity.GenerateTestRequest
		mockSetup      func(*mocks.MockAutotesterAPIService)
		expectedError  bool
		expectedOutput entity.GenerateTestToolResponse
	}{
		{
			name: "successful test generation",
			input: entity.GenerateTestRequest{
				Prompt: "Create a test for login",
				UserId: "user123",
				ChatId: "chat456",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.GenerateTestToolResponse{
					GenerateMsg: &entity.GenerateMessage{
						Id:        "msg1",
						Role:      "assistant",
						Body:      "Generated test code here",
						CreatedAt: time.Now(),
					},
					UserId: "user123",
					ChatId: "chat456",
				}
				m.EXPECT().GenerateTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.GenerateTestToolResponse{
				GenerateMsg: &entity.GenerateMessage{
					Id:   "msg1",
					Role: "assistant",
					Body: "Generated test code here",
				},
				UserId: "user123",
				ChatId: "chat456",
			},
		},
		{
			name: "validation feedback received",
			input: entity.GenerateTestRequest{
				Prompt: "Incomplete prompt",
				UserId: "user123",
				ChatId: "chat456",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.GenerateTestToolResponse{
					ValidateMsg: &entity.ValidateMessage{
						Body: "Please provide more context",
					},
					UserId: "user123",
					ChatId: "chat456",
				}
				m.EXPECT().GenerateTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.GenerateTestToolResponse{
				ValidateMsg: &entity.ValidateMessage{
					Body: "Please provide more context",
				},
				UserId: "user123",
				ChatId: "chat456",
			},
		},
		{
			name: "service returns error",
			input: entity.GenerateTestRequest{
				Prompt: "Invalid prompt",
				UserId: "user789",
				ChatId: "chat999",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				m.EXPECT().GenerateTest(mock.Anything, mock.Anything).
					Return(nil, errors.New("backend service unavailable")).Once()
			},
			expectedError:  true,
			expectedOutput: entity.GenerateTestToolResponse{},
		},
		{
			name: "empty prompt",
			input: entity.GenerateTestRequest{
				Prompt: "",
				UserId: "user000",
				ChatId: "chat000",
			},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.GenerateTestToolResponse{
					GenerateMsg: &entity.GenerateMessage{
						Id:        "msg2",
						Role:      "assistant",
						Body:      "",
						CreatedAt: time.Now(),
					},
					UserId: "user000",
					ChatId: "chat000",
				}
				m.EXPECT().GenerateTest(mock.Anything, mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.GenerateTestToolResponse{
				GenerateMsg: &entity.GenerateMessage{
					Id:   "msg2",
					Role: "assistant",
					Body: "",
				},
				UserId: "user000",
				ChatId: "chat000",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := mocks.NewMockAutotesterAPIService(t)
			test.mockSetup(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			tool, err := NewGenerateTestTool(logger, mockService)
			assert.NoError(t, err)

			ctx := context.Background()
			_, output, err := tool.GenerateTest(ctx, nil, test.input)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput.UserId, output.UserId)
				assert.Equal(t, test.expectedOutput.ChatId, output.ChatId)

				if test.expectedOutput.GenerateMsg != nil {
					assert.NotNil(t, output.GenerateMsg)
					assert.Equal(t, test.expectedOutput.GenerateMsg.Id, output.GenerateMsg.Id)
					assert.Equal(t, test.expectedOutput.GenerateMsg.Role, output.GenerateMsg.Role)
					assert.Equal(t, test.expectedOutput.GenerateMsg.Body, output.GenerateMsg.Body)
				} else {
					assert.Nil(t, output.GenerateMsg)
				}

				if test.expectedOutput.ValidateMsg != nil {
					assert.NotNil(t, output.ValidateMsg)
					assert.Equal(t, test.expectedOutput.ValidateMsg.Body, output.ValidateMsg.Body)
				} else {
					assert.Nil(t, output.ValidateMsg)
				}
			}
		})
	}
}
