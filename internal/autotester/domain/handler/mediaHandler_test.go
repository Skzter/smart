package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
)

//nolint:funlen
func TestHandleGetScreenshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		TestID         string
		ExpectedStatus int
		ExpectedBody   string
		HasScreenshot  bool
		GetURLError    error
		ScreenshotURL  string
	}{
		{
			TestName:       "Valid screenshot request",
			TestID:         "test123",
			ExpectedStatus: http.StatusTemporaryRedirect,
			HasScreenshot:  true,
			ScreenshotURL:  "https://s3.example.com/screenshots/test123.png",
			ExpectedBody:   "",
		},
		{
			TestName:       "Missing testId parameter",
			TestID:         "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"error":"testId is required"}`,
		},
		{
			TestName:       "Screenshot not found",
			TestID:         "test456",
			ExpectedStatus: http.StatusNotFound,
			HasScreenshot:  false,
			ExpectedBody:   `{"error":"no screenshot found for this test"}`,
		},
		{
			TestName:       "Error getting screenshot URL",
			TestID:         "test789",
			ExpectedStatus: http.StatusInternalServerError,
			HasScreenshot:  true,
			GetURLError:    errors.New("S3 connection error"),
			ExpectedBody:   `{"error":"failed to retrieve screenshot"}`,
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
			mockMediaStorageServ := mocks.NewMockMediaStorageService(t)
			mockMetricsServ := sharedMocks.NewMockMetricsService(t)
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

			if test.TestID != "" {
				mockMediaStorageServ.On("HasMedia", mock.Anything, test.TestID).
					Return(test.HasScreenshot, false).Maybe()

				if test.HasScreenshot && test.GetURLError == nil {
					mockMediaStorageServ.On("GetScreenshotUrl", mock.Anything, test.TestID).
						Return(test.ScreenshotURL, nil).Maybe()
				} else if test.HasScreenshot && test.GetURLError != nil {
					mockMediaStorageServ.On("GetScreenshotUrl", mock.Anything, test.TestID).
						Return("", test.GetURLError).Maybe()
				}
			}

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaStorageServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
			)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test/"+test.TestID+"/screenshot", nil)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "testId", Value: test.TestID}}

			controller.HandleGetScreenshot(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s",
					test.ExpectedStatus, rec.Code, rec.Body.String())
			}

			if test.ExpectedStatus == http.StatusTemporaryRedirect {
				location := rec.Header().Get("Location")
				if location != test.ScreenshotURL {
					t.Errorf("Expected redirect to %s, got %s",
						test.ScreenshotURL, location)
				}
			} else if rec.Body.String() != test.ExpectedBody {
				t.Errorf("Expected body %s, got %s",
					test.ExpectedBody, rec.Body.String())
			}
		})
	}
}

//nolint:funlen
func TestHandleGetVideo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		TestID         string
		ExpectedStatus int
		ExpectedBody   string
		HasVideo       bool
		GetURLError    error
		VideoURL       string
	}{
		{
			TestName:       "Valid video request",
			TestID:         "test123",
			ExpectedStatus: http.StatusTemporaryRedirect,
			HasVideo:       true,
			VideoURL:       "https://s3.example.com/videos/test123.webm",
			ExpectedBody:   "",
		},
		{
			TestName:       "Missing testId parameter",
			TestID:         "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"error":"testId is required"}`,
		},
		{
			TestName:       "Video not found",
			TestID:         "test456",
			ExpectedStatus: http.StatusNotFound,
			HasVideo:       false,
			ExpectedBody:   `{"error":"no video found for this test"}`,
		},
		{
			TestName:       "Error getting video URL",
			TestID:         "test789",
			ExpectedStatus: http.StatusInternalServerError,
			HasVideo:       true,
			GetURLError:    errors.New("S3 connection error"),
			ExpectedBody:   `{"error":"failed to retrieve video"}`,
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
			mockMediaStorageServ := mocks.NewMockMediaStorageService(t)
			mockMetricsServ := sharedMocks.NewMockMetricsService(t)
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

			if test.TestID != "" {
				mockMediaStorageServ.On("HasMedia", mock.Anything, test.TestID).
					Return(false, test.HasVideo).Maybe()

				if test.HasVideo && test.GetURLError == nil {
					mockMediaStorageServ.On("GetVideoUrl", mock.Anything, test.TestID).
						Return(test.VideoURL, nil).Maybe()
				} else if test.HasVideo && test.GetURLError != nil {
					mockMediaStorageServ.On("GetVideoUrl", mock.Anything, test.TestID).
						Return("", test.GetURLError).Maybe()
				}
			}

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaStorageServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
			)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test/"+test.TestID+"/video", nil)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "testId", Value: test.TestID}}

			controller.HandleGetVideo(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s",
					test.ExpectedStatus, rec.Code, rec.Body.String())
			}

			if test.ExpectedStatus == http.StatusTemporaryRedirect {
				location := rec.Header().Get("Location")
				if location != test.VideoURL {
					t.Errorf("Expected redirect to %s, got %s",
						test.VideoURL, location)
				}
			} else if rec.Body.String() != test.ExpectedBody {
				t.Errorf("Expected body %s, got %s",
					test.ExpectedBody, rec.Body.String())
			}
		})
	}
}

//nolint:funlen
func TestHandleGetMediaInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		TestName       string
		TestID         string
		ExpectedStatus int
		ExpectedBody   string
		HasScreenshot  bool
		HasVideo       bool
	}{
		{
			TestName:       "Both media types available",
			TestID:         "test123",
			ExpectedStatus: http.StatusOK,
			HasScreenshot:  true,
			HasVideo:       true,
			ExpectedBody:   `{"hasScreenshot":true,"hasVideo":true,"testId":"test123"}`,
		},
		{
			TestName:       "Only screenshot available",
			TestID:         "test456",
			ExpectedStatus: http.StatusOK,
			HasScreenshot:  true,
			HasVideo:       false,
			ExpectedBody:   `{"hasScreenshot":true,"hasVideo":false,"testId":"test456"}`,
		},
		{
			TestName:       "Only video available",
			TestID:         "test789",
			ExpectedStatus: http.StatusOK,
			HasScreenshot:  false,
			HasVideo:       true,
			ExpectedBody:   `{"hasScreenshot":false,"hasVideo":true,"testId":"test789"}`,
		},
		{
			TestName:       "No media available",
			TestID:         "test000",
			ExpectedStatus: http.StatusOK,
			HasScreenshot:  false,
			HasVideo:       false,
			ExpectedBody:   `{"hasScreenshot":false,"hasVideo":false,"testId":"test000"}`,
		},
		{
			TestName:       "Missing testId parameter",
			TestID:         "",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"error":"testId is required"}`,
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
			mockMediaStorageServ := mocks.NewMockMediaStorageService(t)
			mockMetricsServ := sharedMocks.NewMockMetricsService(t)
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

			if test.TestID != "" {
				mockMediaStorageServ.On("HasMedia", mock.Anything, test.TestID).
					Return(test.HasScreenshot, test.HasVideo).Maybe()
			}

			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaStorageServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
			)
			if err != nil {
				t.Fatalf("Controller build failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/test/"+test.TestID+"/media", nil)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Params = gin.Params{{Key: "testId", Value: test.TestID}}

			controller.HandleGetMediaInfo(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s",
					test.ExpectedStatus, rec.Code, rec.Body.String())
			}

			if rec.Body.String() != test.ExpectedBody {
				t.Errorf("Expected body %s, got %s",
					test.ExpectedBody, rec.Body.String())
			}
		})
	}
}
