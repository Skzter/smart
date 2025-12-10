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
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

//nolint:gochecknoglobals // test helper
var PipeOverrideDefault = io.Pipe

//nolint:gochecknoglobals // test helper
var StdCopyOverrideDefault = stdcopy.StdCopy

type fakeConn struct{}

func (fakeConn) Read(p []byte) (int, error)       { return 0, io.EOF }
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

func (c *closeNotifyRecorder) CloseNotify() <-chan bool { return c.done }

// nolint: funlen
func TestHandleLogRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	type testCase struct {
		Name            string
		SetupMocks      func(d *mocks.MockDocker)
		PipeOverride    func()
		StdCopyOverride func()
		CloseOverride   func()
		ExpectedStatus  int
		ExpectedBody    []string
	}

	tests := []testCase{
		{
			Name: "Container not found",
			SetupMocks: func(d *mocks.MockDocker) {
				d.On("GetContainerID", "abc").
					Return("", false).
					Once()
			},
			ExpectedStatus: http.StatusNotFound,
			ExpectedBody:   []string{`Container not found`},
		},
		{
			Name: "Attach fails",
			SetupMocks: func(d *mocks.MockDocker) {
				d.On("GetContainerID", "abc").
					Return("cid123", true).
					Once()

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return((*types.HijackedResponse)(nil), io.ErrUnexpectedEOF).
					Once()
			},
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedBody:   []string{`Failed to attach`},
		},
		{
			Name: "Full pipeline: stdcopy error + container error",
			SetupMocks: func(d *mocks.MockDocker) {
				d.On("GetContainerID", "abc").
					Return("cid123", true).
					Once()

				d.On("AttachToContainer", mock.Anything, "cid123").
					Return(&types.HijackedResponse{
						Reader: bufio.NewReader(strings.NewReader("dummy")),
						Conn:   fakeConn{},
					}, nil).Once()

				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				errCh <- errors.New("container boom")

				var statusRecv <-chan container.WaitResponse = statusCh
				var errRecv <-chan error = errCh

				d.On("WaitContainer", mock.Anything, "cid123").
					Return(statusRecv, errRecv).
					Once()
			},
			PipeOverride: func() {
				PipeFactory = PipeOverrideDefault
			},
			StdCopyOverride: func() {
				StdcopyFunc = func(_, _ io.Writer, _ io.Reader) (int64, error) {
					return 0, errors.New("stdcopy boom")
				}
			},
			ExpectedStatus: http.StatusOK,
			ExpectedBody: []string{
				"event:error",
				"data:{}",
				"event:error",
				"data:container errored",
			},
		},
	}

	for _, test := range tests { //nolint:funlen // table-driven test is allowed
		t.Run(test.Name, func(t *testing.T) {
			PipeFactory = PipeOverrideDefault
			StdcopyFunc = StdCopyOverrideDefault
			CloseFunc = func(c io.Closer) error { return c.Close() }

			mockDocker := mocks.NewMockDocker(t)

			if test.PipeOverride != nil {
				test.PipeOverride()
			}
			if test.StdCopyOverride != nil {
				test.StdCopyOverride()
			}
			if test.CloseOverride != nil {
				test.CloseOverride()
			}

			test.SetupMocks(mockDocker)

			rec := newCloseNotifyRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			req := httptest.NewRequest("GET", "/api/v1/log/abc", nil)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "testId", Value: "abc"}}

			controller, _ := NewAutotesterController(
				logger, cfg,
				mocks.NewMockValidator(t),
				mocks.NewMockGeneratePrompt(t),
				mocks.NewMockTestcaseLocalStorageService(t),
				mockDocker,
				mocks.NewMockChatStorageService(t),
				mocks.NewMockTestcaseStorageService(t),
				mocks.NewMockChatManager(t),
			)

			controller.HandleLogRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Fatalf("expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			body := rec.Body.String()
			for _, expect := range test.ExpectedBody {
				if expect != "" && !strings.Contains(body, expect) {
					t.Fatalf("expected body to contain %q\nactual:\n%s", expect, body)
				}
			}
		})
	}
}
