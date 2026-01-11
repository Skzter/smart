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

// nolint:dupl
func TestHandleSaveLocalRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		SetupMock      func(*mocks.MockTestcaseLocalStorageService)
	}{
		{
			TestName: "Valid save request",
			RequestBody: `{
				"userId": "user123",
				"conversationId": "conv456",
				"code": "import { test, expect } from '@playwright/test';\n\ntest('example test', async ({ page }) => {\n  await page.goto('https://example.com');\n});"
			}`,
			ExpectedStatus: http.StatusOK,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Save(mock.Anything, "user123", "conv456").Return(nil).Once()
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
		},
		{
			TestName: "Save service fails",
			RequestBody: `{
				"userId": "user789",
				"conversationId": "conv789",
				"code": "test code"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Save(mock.Anything, "user789", "conv789").Return(errors.New("database error")).Once()
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
			mockGroups := mocks.NewMockGroupStorage(t)

			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

			test.SetupMock(mockLocalStorageServ)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/saveLocal", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
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
				mockGroups,
				tracer,
				mockMetricsServ,
			)
			if err != nil {
				t.Errorf("build failed")
			}

			controller.HandleSaveLocalRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}

// nolint:dupl,funlen
func TestHandleDeleteLocalRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		QueryParams    map[string]string
		ExpectedStatus int
		SetupMock      func(*mocks.MockTestcaseLocalStorageService)
	}{
		{
			TestName: "Valid delete request",
			QueryParams: map[string]string{
				"testcaseId":     "test123",
				"userId":         "user123",
				"conversationId": "conv456",
			},
			ExpectedStatus: http.StatusOK,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Delete("test123", "user123", "conv456").Return(nil).Once()
			},
		},
		{
			TestName: "Missing required parameters",
			QueryParams: map[string]string{
				"testcaseId": "test123",
			},
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
		},
		{
			TestName: "Delete service fails",
			QueryParams: map[string]string{
				"testcaseId":     "test789",
				"userId":         "user789",
				"conversationId": "conv789",
			},
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Delete("test789", "user789", "conv789").Return(errors.New("database error")).Once()
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
			mockGroups := mocks.NewMockGroupStorage(t)

			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

			test.SetupMock(mockLocalStorageServ)

			url := "/api/v1/deleteLocal"
			if len(test.QueryParams) > 0 {
				url += "?"
				first := true
				for key, value := range test.QueryParams {
					if !first {
						url += "&"
					}
					url += key + "=" + value
					first = false
				}
			}

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

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
				mockGroups,
				tracer,
				mockMetricsServ,
			)
			if err != nil {
				t.Errorf("build failed")
			}

			controller.HandleDeleteLocalRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}
