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

	type mockSetupFunc func(
		mockLocalStorageServ *mocks.MockTestcaseLocalStorageService,
		mockRemoteStorageServ *mocks.MockTestcaseStorageService,
		mockDockerServ *mocks.MockDocker,
	)

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		MockSetup      mockSetupFunc
	}{
		{
			TestName: "Valid run container request",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/session789.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(nil)
				docker.EXPECT().ReadLog(mock.Anything).Return("successful test", nil)
			},
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
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("files not found"))
			},
		},
		{
			TestName: "docker container execution fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/test.spec.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(errors.New("running error"))
			},
		},
		{
			TestName: "Log file read fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/test.spec.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(nil)
				docker.EXPECT().ReadLog(mock.Anything).Return("no logs", errors.New("failed to read log"))
			},
		},
		{
			TestName: "Test passed, Read local testcode fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/test.spec.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(nil)
				docker.EXPECT().ReadLog(mock.Anything).Return("✓ 1 test.spec.ts:12:34", nil)
				local.EXPECT().Read(mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("read failed"))
			},
		},
		{
			TestName: "Test passed, SaveTestCase fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/test.spec.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(nil)
				docker.EXPECT().ReadLog(mock.Anything).Return("✓ 1 test.spec.ts:12:34", nil)
				local.EXPECT().Read(mock.Anything, mock.Anything, mock.Anything).Return("testcode", nil)
				remote.EXPECT().SaveTestcase(mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("save failed"))
			},
		},
		{
			TestName: "Test passed, SaveTestCase succeeds",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, remote *mocks.MockTestcaseStorageService, docker *mocks.MockDocker) {
				local.EXPECT().GetTestPath(mock.Anything, mock.Anything, mock.Anything).Return("/tmp/test.spec.ts", nil)
				docker.EXPECT().RunTest(mock.Anything, mock.Anything).Return(nil)
				docker.EXPECT().ReadLog(mock.Anything).Return("✓ 1 test.spec.ts:12:34", nil)
				local.EXPECT().Read(mock.Anything, mock.Anything, mock.Anything).Return("testcode", nil)
				remote.EXPECT().SaveTestcase(mock.Anything, mock.Anything, mock.Anything).Return("key", nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)

			if test.MockSetup != nil {
				test.MockSetup(mockLocalStorageServ, mockRemoteStorageServ, mockDockerServ)
			}

			req, err := http.NewRequest(http.MethodPost, "/api/v1/run", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			controller.HandleRunContainer(ctx)
			t.Logf("TEST => %v\nRECORDER => %v", test, rec)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", test.ExpectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
