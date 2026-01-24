package service

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/repository"
	mocksStore "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/store"
)

func TestNewAutotesterAPIService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	mockRepo := mocks.NewMockAutotesterAPIRepository(t)
	mockStore := mocksStore.NewMockTestLogStreamStore(t)

	tests := []struct {
		name      string
		logger    *slog.Logger
		repo      *mocks.MockAutotesterAPIRepository
		store     *mocksStore.MockTestLogStreamStore
		expectErr bool
	}{
		{
			name:      "success",
			logger:    logger,
			repo:      mockRepo,
			store:     mockStore,
			expectErr: false,
		},
		{
			name:      "nil-logger",
			logger:    nil,
			repo:      mockRepo,
			store:     mockStore,
			expectErr: true,
		},
		{
			name:      "nil-repo",
			logger:    logger,
			repo:      nil,
			store:     mockStore,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewAutotesterAPIService(test.logger, test.repo, test.store)
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
	logger := slog.New(slog.DiscardHandler)

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
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
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
	logger := slog.New(slog.DiscardHandler)

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
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
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
	logger := slog.New(slog.DiscardHandler)

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
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
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

// nolint:funlen
func TestReadTestLogStream(t *testing.T) {
	tests := []struct {
		name                 string
		testId               string
		expectedStoreContent []entity.LogEvent
		excpectError         bool
		cancelContextDuring  bool
		readLogs             bool
		setupMock            func(*mocks.MockAutotesterAPIRepository, *mocksStore.MockTestLogStreamStore, *[]entity.LogEvent)
	}{
		{
			name:   "success - transmits all events to store",
			testId: "test-123",
			expectedStoreContent: []entity.LogEvent{
				{Event: "log", Data: "test started"},
				{Event: "log", Data: "something"},
				{Event: "finish", Data: "terminated"},
			},
			excpectError: false,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent) {
				events := []entity.LogEvent{
					{Event: "log", Data: "test started"},
					{Event: "log", Data: "something"},
					{Event: "finish", Data: "terminated"},
				}

				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-123", mock.Anything).
					Run(func(ctx context.Context, id string, ch chan<- *entity.LogEvent) {
						for i := range events {
							ch <- &events[i]
						}
					}).
					Return(nil).
					Once()

				ms.EXPECT().
					AddEvent("test-123", mock.Anything).
					Run(func(id string, ev entity.LogEvent) {
						*captured = append(*captured, ev)
					}).
					Return().
					Times(3)

				ms.EXPECT().CompleteStream("test-123").Once()
			},
		},
		{
			name:                 "repo returns error when connecting",
			testId:               "test-123",
			expectedStoreContent: []entity.LogEvent{},
			excpectError:         true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-123", mock.Anything).
					Return(errors.New("connection failed")).
					Once()

				ms.EXPECT().CompleteStream("test-123").Once()
			},
		},
		{
			name:   "handles nil events - filters them out",
			testId: "test-nil",
			expectedStoreContent: []entity.LogEvent{
				{Event: "valid", Data: "data"},
			},
			excpectError: false,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent) {
				mr.EXPECT().ReadTestLogStream(mock.Anything, "test-nil", mock.Anything).
					Run(func(ctx context.Context, id string, ch chan<- *entity.LogEvent) {
						ch <- &entity.LogEvent{Event: "valid", Data: "data"}
						ch <- nil // should be ignored
					}).Return(nil).Once()
				ms.EXPECT().AddEvent("test-nil", mock.Anything).Run(func(id string, ev entity.LogEvent) {
					*captured = append(*captured, ev)
				}).Return().Once()
				ms.EXPECT().CompleteStream("test-nil").Once()
			},
		},
		{
			name:                 "terminates processing on context cancellation",
			testId:               "test-cancel",
			expectedStoreContent: []entity.LogEvent{},
			excpectError:         true,
			cancelContextDuring:  true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-cancel", mock.Anything).
					Run(func(ctx context.Context, id string, ch chan<- *entity.LogEvent) {
						<-ctx.Done()
					}).
					Return(context.Canceled).Once()

				ms.EXPECT().CompleteStream("test-cancel").Once()
			},
		},
		{
			name:   "channel fills to 90% capacity - monitor logs warning",
			testId: "test-full",
			expectedStoreContent: func() []entity.LogEvent {
				events := make([]entity.LogEvent, 1900)
				for i := range 1900 {
					events[i] = entity.LogEvent{Event: "log", Data: "event"}
				}
				return events
			}(),
			excpectError: false,
			readLogs:     true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-full", mock.Anything).
					Run(func(ctx context.Context, id string, ch chan<- *entity.LogEvent) {
						// Send many events rapidly to fill channel to ~90% capacity
						for range 1900 {
							ch <- &entity.LogEvent{Event: "log", Data: "event"}
						}
					}).
					Return(nil).
					Once()

				ms.EXPECT().
					AddEvent("test-full", mock.Anything).
					Run(func(id string, ev entity.LogEvent) {
						*captured = append(*captured, ev)
						// Slow down consumer to allow channel to fill up
						time.Sleep(100 * time.Microsecond)
					}).
					Return().
					Times(1900)

				ms.EXPECT().CompleteStream("test-full").Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			capturedEvents := []entity.LogEvent{}

			test.setupMock(mockRepo, mockStore, &capturedEvents)
			logBuffer := &bytes.Buffer{}
			logger := slog.New(slog.NewTextHandler(logBuffer, nil))
			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)

			require.NoError(t, err)

			var ctx context.Context
			var cancel context.CancelFunc

			if test.cancelContextDuring {
				ctx, cancel = context.WithCancel(context.Background())
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
			}
			defer cancel()

			err = svc.ReadTestLogStream(ctx, test.testId)

			if test.excpectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedStoreContent, capturedEvents)

				if test.readLogs {
					logOutput := logBuffer.String()
					require.Contains(t, logOutput, "Event channel at 90% capacity",
						"expected warning log for channel capacity")
				}
			}
		})
	}
}
