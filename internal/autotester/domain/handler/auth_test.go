package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
)

// nolint:funlen
func TestHandleGenerateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	now := time.Now()
	validToken := &entity.Token{
		UserID:    "user123",
		Token:     "generated-token-abc",
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
		RevokedAt: nil,
	}

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		ExpectedBody   string
		MockSetup      []any
	}{
		{
			TestName:       "Valid generate token request",
			RequestBody:    `{"userId": "user123"}`,
			ExpectedStatus: http.StatusOK,
			ExpectedBody:   mustMarshal(validToken),
			MockSetup:      []any{validToken, nil},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"Bad Request"}`,
		},
		{
			TestName:       "Empty userId",
			RequestBody:    `{"userId": ""}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"Bad Request"}`,
		},
		{
			TestName:       "Missing userId field",
			RequestBody:    `{}`,
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody:   `{"message":"Bad Request"}`,
		},
		{
			TestName:       "GenerateToken service error",
			RequestBody:    `{"userId": "user123"}`,
			ExpectedStatus: http.StatusInternalServerError,
			ExpectedBody:   `{"message":"Internal Server Error"}`,
			MockSetup:      []any{nil, errors.New("database error")},
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
			// Setup metrics mock to accept any calls
			mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
			mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
			mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			if test.MockSetup != nil {
				mockAuth.On("GenerateToken", mock.Anything, mock.Anything).
					Return(test.MockSetup...)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewBufferString(test.RequestBody))
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

			controller.HandleGenerateToken(ctx)

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

// mustMarshal marshals v to JSON string, panics on error.
func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
