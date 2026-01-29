package service

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
)

func TestNewJWTAuthentification(t *testing.T) {
	tests := []struct {
		name      string
		logger    *slog.Logger
		expectErr bool
	}{
		{
			name:      "success",
			logger:    slog.New(slog.DiscardHandler),
			expectErr: false,
		},
		{
			name:      "nil-logger",
			logger:    nil,
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewJWTAuthentification(test.logger)
			if test.expectErr {
				require.Error(t, err)
				require.Nil(t, service)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

func TestJWTExtraction(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	service, err := NewJWTAuthentification(logger)
	require.NoError(t, err)

	tests := []struct {
		name              string
		authHeader        string
		expectedAbort     bool
		expectedAbortCode int
		expectedJWTSet    bool
		expectedJWTString string
	}{
		{
			name:              "success-valid-bearer-token",
			authHeader:        "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedAbort:     false,
			expectedAbortCode: 0,
			expectedJWTSet:    true,
			expectedJWTString: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:              "error-missing-token",
			authHeader:        "",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
		{
			name:              "error-malformed-header-no-bearer",
			authHeader:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
		{
			name:              "error-malformed-header-only-bearer",
			authHeader:        "Bearer",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
		{
			name:              "error-wrong-scheme",
			authHeader:        "Basic dXNlcjpwYXNz",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
		{
			name:              "error-empty-bearer-token-with-space",
			authHeader:        "Bearer ",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
		{
			name:              "error-bearer-with-only-whitespace",
			authHeader:        "Bearer    ",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
			expectedJWTString: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if test.authHeader != "" {
				req.Header.Set("Authorization", test.authHeader)
			}
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req
			handler := service.JWTExtractionIntoContext()

			handler(ctx)

			if test.expectedAbort {
				require.Equal(t, test.expectedAbortCode, w.Code)
			} else {
				require.Equal(t, http.StatusOK, w.Code)

				if test.expectedJWTSet {
					val := ctx.Request.Context().Value(entity.JwtContextKey{})
					require.NotNil(t, val)

					jwtValue, ok := val.(string)
					require.True(t, ok)
					require.Equal(t, test.expectedJWTString, jwtValue)
				}
			}
		})
	}
}
