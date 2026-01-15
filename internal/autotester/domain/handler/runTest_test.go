package handler

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"

	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint:funlen
func TestHandleRunContainer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	type mockSetupFunc func(
		local *mocks.MockTestcaseLocalStorageService,
		docker *mocks.MockDocker,
	)

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		ExpectedBody   string
		MockSetup      mockSetupFunc
	}{
		{
			TestName: "Valid run container request",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"chatId": "chat789"
			}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   `{"result":"Test started"}`,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, docker *mocks.MockDocker) {
				local.On("GetTestPath", "test456", "user123", "chat789").
					Return("/tmp/chat789.ts", nil)

				docker.On("RunTest",
					mock.Anything,
					"/tmp/chat789.ts",
					"test456",
					"user123",
					"chat789",
				).Return("container-id", nil)
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"Bad Request"}`,
		},
		{
			TestName: "Missing required fields",
			RequestBody: `{
				"testId": "test456",
				"chatId": "chat789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"Missing required parameters"}`,
		},
		{
			TestName: "GetTestPath fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"chatId": "chat789"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"files not found"}`,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, docker *mocks.MockDocker) {
				local.On("GetTestPath", "test456", "user123", "chat789").
					Return("", errors.New("files not found"))
			},
		},
		{
			TestName: "RunTest fails",
			RequestBody: `{
				"userId": "user123",
				"testId": "test456",
				"chatId": "chat789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedBody:   `{"message":"running error"}`,
			MockSetup: func(local *mocks.MockTestcaseLocalStorageService, docker *mocks.MockDocker) {
				local.On("GetTestPath", "test456", "user123", "chat789").
					Return("/tmp/test.spec.ts", nil)

				docker.On("RunTest",
					mock.Anything,
					"/tmp/test.spec.ts",
					"test456",
					"user123",
					"chat789",
				).Return("", errors.New("running error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatManager := mocks.NewMockChatManager(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockMetricsServ := sharedMocks.NewMockMetricsService(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()

			if test.MockSetup != nil {
				test.MockSetup(mockLocalStorageServ, mockDockerServ)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/run", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
			)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			controller.HandleRunContainer(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s",
					test.ExpectedStatus, rec.Code, rec.Body.String())
			}

			if rec.Body.String() != test.ExpectedBody {
				t.Errorf("Expected body %s, got %s",
					test.ExpectedBody, rec.Body.String())
			}

			mockLocalStorageServ.AssertExpectations(t)
			mockDockerServ.AssertExpectations(t)
		})
	}
}
