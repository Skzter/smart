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

const token = "token"

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
		token           string
		setupMock       func(*mocks.MockAutotesterAPIRepository, string)
		expectErr       bool
		expectedContent string
		ctx             context.Context
	}{
		{
			name:  "success",
			token: token,
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					GetTemplate(mock.Anything, token).
					Return(&entity.TemplateResponse{Content: "test template"}, nil).
					Once()
			},
			expectErr:       false,
			expectedContent: "test template",
			ctx:             context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "repository-error",
			token: token,
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					GetTemplate(mock.Anything, token).
					Return(nil, errors.New("repo error")).
					Once()
			},
			expectErr:       true,
			expectedContent: "",
			ctx:             context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "empty-token",
			token: "",
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					GetTemplate(mock.Anything, token).
					Return(nil, errors.New("unauthorized")).
					Once()
			},
			expectErr:       true,
			expectedContent: "",
			ctx:             context.WithValue(context.Background(), entity.JwtContextKey{}, ""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo, test.token)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
			require.NoError(t, err)

			res, err := svc.GetTemplate(test.ctx)
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
		token               string
		request             *entity.GenerateTestRequest
		setupMock           func(*mocks.MockAutotesterAPIRepository, string)
		expectErr           bool
		expectedValidateMsg *entity.ValidateMessage
		expectedGenerateMsg *entity.GenerateMessage
		ctx                 context.Context
	}{
		{
			name:  "success - validation passes, test generated",
			token: token,
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					ValidatePrompt(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(&entity.ValidatePromptResponse{
					Result: entity.ValidateMessage{
						Body: "",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
				m.EXPECT().
					GenerateTest(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(&entity.GenerateTestResponse{
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
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "validation fails - returns validation message",
			token: token,
			request: &entity.GenerateTestRequest{
				Prompt: "bad prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					ValidatePrompt(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "bad prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(&entity.ValidatePromptResponse{
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
			ctx:                 context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:    "nil-request",
			token:   token,
			request: nil,
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
			ctx:                 context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "validation-error",
			token: token,
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					ValidatePrompt(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(nil, errors.New("validation error")).Once()
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
			ctx:                 context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "generate-error",
			token: token,
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					ValidatePrompt(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(&entity.ValidatePromptResponse{
					Result: entity.ValidateMessage{
						Body: "",
					},
					UserId: "user-123",
					ChatId: "chat-456",
				}, nil).Once()
				m.EXPECT().
					GenerateTest(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(nil, errors.New("generate error")).Once()
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
			ctx:                 context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "empty-token",
			token: "",
			request: &entity.GenerateTestRequest{
				Prompt: "test prompt",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					ValidatePrompt(mock.Anything, &entity.GenerateTestRequest{
						Prompt: "test prompt",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).Return(nil, errors.New("unauthorized")).Once()
			},
			expectErr:           true,
			expectedValidateMsg: nil,
			expectedGenerateMsg: nil,
			ctx:                 context.WithValue(context.Background(), entity.JwtContextKey{}, ""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo, test.token)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
			require.NoError(t, err)

			res, err := svc.GenerateTest(test.ctx, test.request)
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
		token         string
		request       *entity.ExecuteTestRequest
		setupMock     func(*mocks.MockAutotesterAPIRepository, string)
		expectErr     bool
		expectedMatch string
		ctx           context.Context
	}{
		{
			name:  "success",
			token: token,
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					SaveTest(mock.Anything, &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(&entity.SaveTestResponse{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						Action: "saved",
					}, nil).
					Once()

				m.EXPECT().
					RunTest(mock.Anything, &entity.RunTestRequest{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(&entity.RunTestResponse{
						Result: "passed",
					}, nil).
					Once()
			},
			expectErr:     false,
			expectedMatch: "saved:testId=550e8400-e29b-41d4-a716-446655440000 action=saved; runResult=passed",
			ctx:           context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:    "nil-request",
			token:   token,
			request: nil,
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
			},
			expectErr:     true,
			expectedMatch: "",
			ctx:           context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "save-error",
			token: token,
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					SaveTest(mock.Anything, &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(nil, errors.New("save error")).
					Once()
			},
			expectErr:     true,
			expectedMatch: "",
			ctx:           context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "run-error",
			token: token,
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					SaveTest(mock.Anything, &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(&entity.SaveTestResponse{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						Action: "saved",
					}, nil).
					Once()

				m.EXPECT().
					RunTest(mock.Anything, &entity.RunTestRequest{
						TestId: "550e8400-e29b-41d4-a716-446655440000",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(nil, errors.New("run error")).
					Once()
			},
			expectErr:     true,
			expectedMatch: "",
			ctx:           context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:  "empty-token",
			token: "",
			request: &entity.ExecuteTestRequest{
				Test:   "test code",
				UserId: "user-123",
				ChatId: "chat-456",
			},
			setupMock: func(m *mocks.MockAutotesterAPIRepository, token string) {
				m.EXPECT().
					SaveTest(mock.Anything, &entity.SaveTestRequest{
						Code:   "test code",
						UserId: "user-123",
						ChatId: "chat-456",
					}, token).
					Return(nil, errors.New("unauthorized")).
					Once()
			},
			expectErr:     true,
			expectedMatch: "",
			ctx:           context.WithValue(context.Background(), entity.JwtContextKey{}, ""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			test.setupMock(mockRepo, test.token)

			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)
			require.NoError(t, err)

			res, err := svc.ExecuteTest(test.ctx, test.request)
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
		token                string
		testId               string
		expectedStoreContent []entity.LogEvent
		excpectError         bool
		cancelContextDuring  bool
		readLogs             bool
		setupMock            func(*mocks.MockAutotesterAPIRepository, *mocksStore.MockTestLogStreamStore, *[]entity.LogEvent, string)
		ctx                  context.Context
	}{
		{
			name:   "success - transmits all events to store",
			token:  token,
			testId: "test-123",
			expectedStoreContent: []entity.LogEvent{
				{Event: "log", Data: "test started"},
				{Event: "log", Data: "something"},
				{Event: "finish", Data: "terminated"},
			},
			excpectError: false,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				events := []entity.LogEvent{
					{Event: "log", Data: "test started"},
					{Event: "log", Data: "something"},
					{Event: "finish", Data: "terminated"},
				}

				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-123", token, mock.Anything).
					Run(func(ctx context.Context, id, token string, ch chan<- *entity.LogEvent) {
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
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:                 "repo returns error when connecting",
			token:                token,
			testId:               "test-123",
			expectedStoreContent: []entity.LogEvent{},
			excpectError:         true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-123", token, mock.Anything).
					Return(errors.New("connection failed")).
					Once()

				ms.EXPECT().CompleteStream("test-123").Once()
			},
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:   "handles nil events - filters them out",
			token:  token,
			testId: "test-nil",
			expectedStoreContent: []entity.LogEvent{
				{Event: "valid", Data: "data"},
			},
			excpectError: false,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				mr.EXPECT().ReadTestLogStream(mock.Anything, "test-nil", token, mock.Anything).
					Run(func(ctx context.Context, id, token string, ch chan<- *entity.LogEvent) {
						ch <- &entity.LogEvent{Event: "valid", Data: "data"}
						ch <- nil // should be ignored
					}).Return(nil).Once()
				ms.EXPECT().AddEvent("test-nil", mock.Anything).Run(func(id string, ev entity.LogEvent) {
					*captured = append(*captured, ev)
				}).Return().Once()
				ms.EXPECT().CompleteStream("test-nil").Once()
			},
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:                 "terminates processing on context cancellation",
			token:                token,
			testId:               "test-cancel",
			expectedStoreContent: []entity.LogEvent{},
			excpectError:         true,
			cancelContextDuring:  true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-cancel", token, mock.Anything).
					Run(func(ctx context.Context, id, token string, ch chan<- *entity.LogEvent) {
						<-ctx.Done()
					}).
					Return(context.Canceled).Once()

				ms.EXPECT().CompleteStream("test-cancel").Once()
			},
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:   "channel fills to 90% capacity - monitor logs warning",
			token:  token,
			testId: "test-full",
			expectedStoreContent: func() []entity.LogEvent {
				events := make([]entity.LogEvent, 2150)
				for i := range 2150 {
					events[i] = entity.LogEvent{Event: "log", Data: "event"}
				}
				return events
			}(),
			excpectError: false,
			readLogs:     true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				release := make(chan struct{})
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-full", token, mock.Anything).
					Run(func(ctx context.Context, id, token string, ch chan<- *entity.LogEvent) {
						// Send many events rapidly to fill channel to ~90% capacity
						for range 1900 {
							ch <- &entity.LogEvent{Event: "log", Data: "event"}
						}
						close(release)
						for range 250 {
							ch <- &entity.LogEvent{Event: "log", Data: "event"}
						}
					}).
					Return(nil).
					Once()

				ms.EXPECT().
					AddEvent("test-full", mock.Anything).
					Run(func(id string, ev entity.LogEvent) {
						*captured = append(*captured, ev)
						<-release // simulates blocked consumer
					}).
					Return().
					Times(2150)

				ms.EXPECT().CompleteStream("test-full").Once()
			},
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, token),
		},
		{
			name:                 "empty-token",
			token:                "",
			testId:               "test-token",
			expectedStoreContent: []entity.LogEvent{},
			excpectError:         true,
			setupMock: func(mr *mocks.MockAutotesterAPIRepository, ms *mocksStore.MockTestLogStreamStore, captured *[]entity.LogEvent, token string) {
				mr.EXPECT().
					ReadTestLogStream(mock.Anything, "test-token", token, mock.Anything).
					Return(errors.New("unauthorized")).
					Once()

				ms.EXPECT().CompleteStream("test-token").Once()
			},
			ctx: context.WithValue(context.Background(), entity.JwtContextKey{}, ""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockAutotesterAPIRepository(t)
			mockStore := mocksStore.NewMockTestLogStreamStore(t)
			capturedEvents := []entity.LogEvent{}

			test.setupMock(mockRepo, mockStore, &capturedEvents, test.token)
			logBuffer := &bytes.Buffer{}
			logger := slog.New(slog.NewTextHandler(logBuffer, nil))
			svc, err := NewAutotesterAPIService(logger, mockRepo, mockStore)

			require.NoError(t, err)

			var ctx context.Context
			var cancel context.CancelFunc

			if test.cancelContextDuring {
				ctx, cancel = context.WithCancel(test.ctx)
				go func() {
					time.Sleep(50 * time.Millisecond)
					cancel()
				}()
			} else {
				ctx, cancel = context.WithTimeout(test.ctx, 2*time.Second)
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
