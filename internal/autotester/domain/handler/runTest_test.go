package handler

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service/mocks"
)

// nolint:funlen
func TestHandleRunContainer(t *testing.T) {
	t.Skip("Skipping test for HandleRunContainer because its currently broken and should not call os commands in tests")
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		SetupMock      func(*mocks.MockTestcaseLocalStorageService)
		SetupTest      func() // For file system operations
		CleanupTest    func() // For cleanup operations
	}{
		{
			TestName: "Valid run container request",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusOK,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().GetTestPath("test456", "user123", "session789").
					Return("/path/to/test.spec.ts", nil).Once()
			},
			SetupTest: func() {
				err := os.MkdirAll(cfg.LogDirAutopw, 0750)
				if err != nil {
					t.Log(err)
				}
				err = os.WriteFile(cfg.LogDirAutopw+"test.spec.ts.log", []byte("Test execution successful"), 0600)
				if err != nil {
					t.Log(err)
				}
			},
			CleanupTest: func() {
				err := os.Remove(cfg.LogDirAutopw + "test.spec.ts.log")
				if err != nil {
					t.Log(err)
				}
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName: "Missing userId field",
			RequestBody: `{
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName: "Missing testId field",
			RequestBody: `{
				"userId": "user123",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName: "Missing sessionId field",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName:       "Empty request body",
			RequestBody:    `{}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName: "All fields empty strings",
			RequestBody: `{
				"userId": "",
				"testId": "",
				"sessionId": ""
			}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
			SetupTest:      func() {},
			CleanupTest:    func() {},
		},
		{
			TestName: "GetTestPath fails - file not found",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().GetTestPath("test456", "user123", "session789").
					Return("", errors.New("test file not found")).Once()
			},
			SetupTest:   func() {},
			CleanupTest: func() {},
		},
		{
			TestName: "Command execution fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().GetTestPath("test456", "user123", "session789").
					Return("/path/to/test.spec.ts", nil).Once()
			},
			SetupTest:   func() {},
			CleanupTest: func() {},
		},
		{
			TestName: "Log file read fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"sessionId": "session789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().GetTestPath("test456", "user123", "session789").
					Return("/path/to/test.spec.ts", nil).Once()
			},
			SetupTest: func() {
				err := os.Remove("docker/logs/output.log")
				if err != nil {
					t.Log(err)
				}
			},
			CleanupTest: func() {},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			// Setup
			test.SetupTest()
			defer test.CleanupTest()

			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidatePrompt(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			test.SetupMock(mockLocalStorageServ)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/run", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			// Execute
			controller.HandleRunContainer(ctx)

			// Assert
			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", test.ExpectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
