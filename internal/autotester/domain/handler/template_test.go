package handler

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

func TestHandleTemplate(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		template       string
		ctx            context.Context
		expectedStatus int
	}{
		{
			TestName:       "Request, not empty Template",
			template:       "valid template",
			ctx:            context.Background(),
			expectedStatus: http.StatusOK,
		}, {
			TestName:       "Request, empty Template",
			template:       "",
			ctx:            context.Background(),
			expectedStatus: http.StatusTeapot,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatManager := mocks.NewMockChatManager(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/v1/template", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Errors.Errors()

			cfg.Template = test.template
			mockMediaServ := mocks.NewMockMediaStorageService(t)
			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaServ,
				mockChatManager,
				tracer,
				mockMetricsServ,
			)

			if err != nil {
				t.Errorf("build failed")
			}
			controller.HandleGetTemplate(ctx)

			if rec.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, rec.Code)
			}
		})
	}
}
