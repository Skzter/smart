package handler

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint:funlen
func TestHandleRunContainer(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		TestName         string
		RequestBody      string
		MockResponseFile []any
		MockResponseRun  []any
		MockResponseRead []any
		ExpectedStatus   int
	}{
		{
			TestName: "Valid run container request",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus:   http.StatusOK,
			MockResponseFile: []any{"/tmp/session789.ts", nil},
			MockResponseRun:  []any{nil},
			MockResponseRead: []any{"successful test", nil},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "Missing userId field",
			RequestBody: `{
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "Missing testId field",
			RequestBody: `{
				"userId": "user123",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "Missing sessionId field",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName:       "Empty request body",
			RequestBody:    `{}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "All fields empty strings",
			RequestBody: `{
				"userId": "",
				"testId": "",
				"sessionId": ""
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "GetTestPath fails - file not found",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus:   http.StatusBadRequest,
			MockResponseFile: []any{"", errors.New("files not found")},
		},
		{
			TestName: "docker container execution fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus:   http.StatusInternalServerError,
			MockResponseFile: []any{"/tmp/test.spec.ts", nil},
			MockResponseRun:  []any{errors.New("running error")},
		},
		{
			TestName: "Log file read fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus:   http.StatusInternalServerError,
			MockResponseFile: []any{"/tmp/test.spec.ts", nil},
			MockResponseRun:  []any{nil},
			MockResponseRead: []any{"no logs", errors.New("failed to read log")},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)

			// mock setup
			if test.MockResponseFile != nil {
				mockLocalStorageServ.On("GetTestPath", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseFile...)
			}
			if test.MockResponseRun != nil {
				mockDockerServ.On("RunTest", mock.Anything, mock.Anything).Return(test.MockResponseRun...)
			}
			if test.MockResponseRead != nil {
				mockDockerServ.On("ReadLog", mock.Anything).Return(test.MockResponseRead...)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/run", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			// Execute
			controller.HandleRunContainer(ctx)
			t.Logf("TEST => %v\nRECORDER => %v", test, rec)

			// Assert
			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", test.ExpectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
