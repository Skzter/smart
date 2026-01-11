package handler

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"log/slog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
)

type fakeConn struct{}

func (fakeConn) Read([]byte) (int, error)         { return 0, io.EOF }
func (fakeConn) Write(p []byte) (int, error)      { return len(p), nil }
func (fakeConn) Close() error                     { return nil }
func (fakeConn) LocalAddr() net.Addr              { return fakeAddr("local") }
func (fakeConn) RemoteAddr() net.Addr             { return fakeAddr("remote") }
func (fakeConn) SetDeadline(time.Time) error      { return nil }
func (fakeConn) SetReadDeadline(time.Time) error  { return nil }
func (fakeConn) SetWriteDeadline(time.Time) error { return nil }

type fakeAddr string

func (a fakeAddr) Network() string { return string(a) }
func (a fakeAddr) String() string  { return string(a) }

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
	done chan bool
}

func newCloseNotifyRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		done:             make(chan bool, 1),
	}
}

func (c *closeNotifyRecorder) CloseNotify() <-chan bool {
	return c.done
}

// nolint: funlen
func TestHandleLogRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	type testCase struct {
		name           string
		setupDocker    func(d *mocks.MockDocker)
		setupMocks     func(mockLocal *mocks.MockTestcaseLocalStorageService, mockRemote *mocks.MockTestcaseStorageService)
		expectedStatus int
		expectedBody   []string
	}

	tests := []testCase{
		{
			name: "Container not found",
			setupDocker: func(d *mocks.MockDocker) {
				d.On("GetContainerInfo", "abc").
					Return((*entity.ContainerInfo)(nil), false)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   []string{"Container not found"},
		},
		{
			name: "Attach fails",
			setupDocker: func(d *mocks.MockDocker) {
				info := &entity.ContainerInfo{ContainerID: "cid123"}
				d.On("GetContainerInfo", "abc").
					Return(info, true)

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return((*types.HijackedResponse)(nil), errors.New("attach failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   []string{"Failed to attach"},
		},
		{
			name: "Container exits successfully and testcase is stored",
			setupDocker: func(d *mocks.MockDocker) {
				info := &entity.ContainerInfo{
					ContainerID: "cid123",
					UserID:      "user1",
					SessionID:   "sess1",
				}

				d.On("GetContainerInfo", "abc").
					Return(info, true)

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return(&types.HijackedResponse{
						Reader: bufio.NewReader(strings.NewReader("log line\n")),
						Conn:   fakeConn{},
					}, nil)

				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error)

				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)

				d.On("WaitContainer", mock.Anything, "cid123").
					Return(
						(<-chan container.WaitResponse)(statusCh),
						(<-chan error)(errCh),
					)
			},
			setupMocks: func(mockLocal *mocks.MockTestcaseLocalStorageService, mockRemote *mocks.MockTestcaseStorageService) {
				mockLocal.
					On("Read", "abc", "user1", "sess1").
					Return("test code", nil).
					Once()

				mockRemote.
					On("SaveTestcase", mock.Anything, mock.Anything, "user1").
					Return("testcase-id-123", nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"event:status",
				"event:finished",
			},
		},
		{
			name: "Container exits successfully but testcase cannot be stored",
			setupDocker: func(d *mocks.MockDocker) {
				info := &entity.ContainerInfo{
					ContainerID: "cid123",
					UserID:      "user1",
					SessionID:   "sess1",
				}

				d.On("GetContainerInfo", "abc").
					Return(info, true)

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return(&types.HijackedResponse{
						Reader: bufio.NewReader(strings.NewReader("log line\n")),
						Conn:   fakeConn{},
					}, nil)

				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error)

				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)

				d.On("WaitContainer", mock.Anything, "cid123").
					Return(
						(<-chan container.WaitResponse)(statusCh),
						(<-chan error)(errCh),
					)
			},
			setupMocks: func(
				mockLocal *mocks.MockTestcaseLocalStorageService,
				mockRemote *mocks.MockTestcaseStorageService,
			) {
				mockLocal.
					On("Read", "abc", "user1", "sess1").
					Return("test code", nil).
					Once()

				mockRemote.
					On("SaveTestcase", mock.Anything, mock.Anything, "user1").
					Return("", errors.New("remote storage down")).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"event:status",
				"event:finished",
			},
		},
		{
			name: "Container error is streamed via SSE",
			setupDocker: func(d *mocks.MockDocker) {
				info := &entity.ContainerInfo{ContainerID: "cid123"}
				d.On("GetContainerInfo", "abc").
					Return(info, true)

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return(&types.HijackedResponse{
						Reader: bufio.NewReader(strings.NewReader("dummy")),
						Conn:   fakeConn{},
					}, nil)

				statusCh := make(chan container.WaitResponse)
				errCh := make(chan error, 1)
				errCh <- errors.New("container crashed")

				d.On("WaitContainer", mock.Anything, "cid123").
					Return(
						(<-chan container.WaitResponse)(statusCh),
						(<-chan error)(errCh),
					)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []string{
				"event:error",
				"container errored",
			},
		},
	}

	mockGen := mocks.NewMockGeneratePrompt(t)
	mockVal := mocks.NewMockValidator(t)
	mockLocal := mocks.NewMockTestcaseLocalStorageService(t)
	mockChat := mocks.NewMockChatStorageService(t)
	mockRemote := mocks.NewMockTestcaseStorageService(t)
	mockGroups := mocks.NewMockGroupStorage(t)
	mockChatManager := mocks.NewMockChatManager(t)
	mockMetrics := sharedMocks.NewMockMetricsService(t)

	mockMetrics.On("IncRequestSuccess").Maybe()
	mockMetrics.On("IncRequestError", mock.Anything).Maybe()
	mockMetrics.On("RecordRequestDuration", mock.Anything).Maybe()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDocker := mocks.NewMockDocker(t)
			tc.setupDocker(mockDocker)
			if tc.setupMocks != nil {
				tc.setupMocks(mockLocal, mockRemote)
			}

			rec := newCloseNotifyRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			req := httptest.NewRequest("GET", "/api/v1/log/abc", nil)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "testId", Value: "abc"}}

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockVal,
				mockGen,
				mockLocal,
				mockDocker,
				mockChat,
				mockRemote,
				mockChatManager,
				mockGroups,
				tracer,
				mockMetrics,
			)
			if err != nil {
				t.Fatalf("controller init failed: %v", err)
			}

			controller.HandleLogRequest(ctx)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			body := rec.Body.String()
			for _, expect := range tc.expectedBody {
				if !strings.Contains(body, expect) {
					t.Fatalf(
						"expected body to contain %q\nactual body:\n%s",
						expect,
						body,
					)
				}
			}
		})
	}
}
func TestSafeSend(t *testing.T) {
	tests := []struct {
		name        string
		setupChan   func() chan entity.SSEvent
		expectError bool
		errContains string
	}{
		{
			name: "channel closed - panic is recovered",
			setupChan: func() chan entity.SSEvent {
				ch := make(chan entity.SSEvent)
				close(ch)
				return ch
			},
			expectError: true,
			errContains: "panic while sending event",
		},
		{
			name: "channel blocked",
			setupChan: func() chan entity.SSEvent {
				return make(chan entity.SSEvent) // unbuffered, no receiver
			},
			expectError: true,
			errContains: "failed to send SSEvent",
		},
		{
			name: "channel send succeeds",
			setupChan: func() chan entity.SSEvent {
				return make(chan entity.SSEvent, 1) // buffered
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.setupChan()

			err := safeSend(ch, entity.SSEvent{Type: "log"})

			if tt.expectError {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
