package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/repository"
)

func TestNewAutotesterAPIService(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockAutotesterAPIRepository(t)

	tests := []struct {
		name      string
		logger    *slog.Logger
		repo      *mocks.MockAutotesterAPIRepository
		expectErr bool
	}{
		{
			name:      "success",
			logger:    logger,
			repo:      mockRepo,
			expectErr: false,
		},
		{
			name:      "nil-logger",
			logger:    nil,
			repo:      mockRepo,
			expectErr: true,
		},
		{
			name:      "nil-repo",
			logger:    logger,
			repo:      nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewAutotesterAPIService(test.logger, test.repo)
			if test.expectErr {
				require.Error(t, err)
				require.Nil(t, svc)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, svc)
		})
	}
}

func TestGetTemplate(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name            string
		setupMock       func(*mocks.MockAutotesterAPIRepository)
		expectErr       bool
		expectedContent string
	}{
		{
			name: "success",
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					GetTemplate(context.Background()).
					Return(&entity.TemplateResponse{Content: "test template"}, nil).
					Once()
			},
			expectErr:       false,
			expectedContent: "test template",
		},
		{
			name: "repository-error",
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					GetTemplate(context.Background()).
					Return(nil, errors.New("repo error")).
					Once()
			},
			expectErr:       true,
			expectedContent: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo)
			require.NoError(t, err)

			res, err := svc.GetTemplate(context.Background())
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedContent, res.Content)
		})
	}
}

// nolint:funlen
func TestGenerateTest(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name                string
		request             *entity.GenerateTestRequest
		setupMock           func(*mocks.MockAutotesterAPIRepository)
		expectErr           bool
		expectedValidateMsg *entity.ValidateMessage
		expectedGenerateMsg *entity.GenerateMessage
	}{
		{
			name: "success - validation passes, test generated",
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					ValidatePrompt(context.Background(), &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(&entity.ValidatePromptResponse{
					Result: entity.ValidateMessage{
						Body: "",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
				m.EXPECT().
					GenerateTest(context.Background(), &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(&entity.GenerateTestResponse{
					Result: entity.GenerateMessage{
						Id:   "msg-1",
						Role: "assistant",
						Body: "generated test code",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
			},
			expectErr:           false,
			expectedValidateMsg: nil,
			expectedGenerateMsg: &entity.GenerateMessage{
				Id:   "msg-1",
				Role: "assistant",
				Body: "generated test code",
			},
		},
		{
			name: "validation fails - returns validation message",
			request: &entity.GenerateTestRequest{
				Prompt: "bad prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					ValidatePrompt(context.Background(), &entity.GenerateTestRequest{
						Prompt: "bad prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(&entity.ValidatePromptResponse{
					Result: entity.ValidateMessage{
						Body: "Please provide more details in your prompt",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
			},
			expectErr: false,
			expectedValidateMsg: &entity.ValidateMessage{
				Body: "Please provide more details in your prompt",
			},
			expectedGenerateMsg: nil,
		},
		{
			name:    "nil-request",
			request: nil,
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
		},
		{
			name: "validation-error",
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					ValidatePrompt(context.Background(), &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(nil, errors.New("validation error")).Once()
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
		},
		{
			name: "generate-error",
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					ValidatePrompt(context.Background(), &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(&entity.ValidatePromptResponse{
					Result: entity.ValidateMessage{
						Body: "",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
				m.EXPECT().
					GenerateTest(context.Background(), &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}).Return(nil, errors.New("generate error")).Once()
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo)
			require.NoError(t, err)

			res, err := svc.GenerateTest(context.Background(), test.request)
			if test.expectErr {
				require.Error(t, err)
				require.Nil(t, res)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedValidateMsg, res.ValidateMsg)
			require.Equal(t, test.expectedGenerateMsg, res.GenerateMsg)
		})
	}
}

// nolint:funlen
func TestExecuteTest(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name          string
		request       *entity.ExecuteTestRequest
		setupMock     func(*mocks.MockAutotesterAPIRepository)
		expectErr     bool
		expectedMatch string
	}{
		{
			name: "success",
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					SaveTest(context.Background(), &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}).
					Return(&entity.SaveTestResponse{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						Action: "saved",
					}, nil).
					Once()

				m.EXPECT().
					RunTest(context.Background(), &entity.RunTestRequest{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						UserId: "user-123",
						ChatId: "chat-456",
					}).
					Return(&entity.RunTestResponse{
						Result: "passed",
					}, nil).
					Once()
			},
			expectErr:     false,
			expectedMatch: "saved:testId=550e8400-e29b-41d4-a716-446655440000 action=saved; runResult=passed",
		},
		{
			name:    "nil-request",
			request: nil,
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
			},
			expectErr:     true,
			expectedMatch: "",
		},
		{
			name: "save-error",
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					SaveTest(context.Background(), &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}).
					Return(nil, errors.New("save error")).
					Once()
			},
			expectErr:     true,
			expectedMatch: "",
		},
		{
			name: "run-error",
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository) {
				m.EXPECT().
					SaveTest(context.Background(), &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}).
					Return(&entity.SaveTestResponse{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						Action: "saved",
					}, nil).
					Once()

				m.EXPECT().
					RunTest(context.Background(), &entity.RunTestRequest{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						UserId: "user-123",
						ChatId: "chat-456",
					}).
					Return(nil, errors.New("run error")).
					Once()
			},
			expectErr:     true,
			expectedMatch: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo)
			require.NoError(t, err)

			res, err := svc.ExecuteTest(context.Background(), test.request)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			require.Equal(t, test.expectedMatch, res.Result)
		})
	}
}
