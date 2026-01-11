package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint:funlen
func TestHandleValidateJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(m *mocks.MockJWTValidator)
		expectedStatus int
		expectValidate bool
	}{
		{
			name:           "missing Authorization header -> unauthorized",
			authHeader:     "",
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectValidate: false,
		},
		{
			name:           "wrong scheme -> unauthorized",
			authHeader:     "Token abc",
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectValidate: false,
		},
		{
			name:           "empty bearer token -> unauthorized",
			authHeader:     "Bearer    ",
			setupMock:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectValidate: false,
		},
		{
			name:       "Validate returns error -> unauthorized",
			authHeader: "Bearer abc",
			setupMock: func(m *mocks.MockJWTValidator) {
				m.On("Validate", mock.Anything, "abc").
					Return(entity.ValidationResult{Valid: false, Revoked: false}, errors.New("boom")).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectValidate: true,
		},
		{
			name:       "token invalid -> unauthorized",
			authHeader: "Bearer abc",
			setupMock: func(m *mocks.MockJWTValidator) {
				m.On("Validate", mock.Anything, "abc").
					Return(entity.ValidationResult{Valid: false, Revoked: false}, nil).
					Once()
			},
			expectedStatus: http.StatusUnauthorized,
			expectValidate: true,
		},
		{
			name:       "token revoked -> forbidden",
			authHeader: "Bearer abc",
			setupMock: func(m *mocks.MockJWTValidator) {
				m.On("Validate", mock.Anything, "abc").
					Return(entity.ValidationResult{Valid: true, Revoked: true}, nil).
					Once()
			},
			expectedStatus: http.StatusForbidden,
			expectValidate: true,
		},
		{
			name:       "token valid -> ok",
			authHeader: "Bearer abc",
			setupMock: func(m *mocks.MockJWTValidator) {
				m.On("Validate", mock.Anything, "abc").
					Return(entity.ValidationResult{Valid: true, Revoked: false}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectValidate: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockJWTValidator := mocks.NewMockJWTValidator(t)

			controller := &AutotesterController{
				logger:       logger,
				jwtValidator: mockJWTValidator,
			}

			if tc.setupMock != nil {
				tc.setupMock(mockJWTValidator)
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/validate", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller.HandleValidateJWT(ctx)

			assert.Equal(t, tc.expectedStatus, rec.Code)

			if !tc.expectValidate {
				mockJWTValidator.AssertNotCalled(t, "Validate", mock.Anything, mock.Anything)
			}
		})
	}
}
