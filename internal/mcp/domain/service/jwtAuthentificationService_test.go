package service

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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
	}{
		{
			name:              "success-valid-bearer-token",
			authHeader:        "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedAbort:     false,
			expectedAbortCode: 0,
			expectedJWTSet:    true,
		},
		{
			name:              "error-missing-token",
			authHeader:        "",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
		},
		{
			name:              "error-malformed-header-no-bearer",
			authHeader:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
		},
		{
			name:              "error-malformed-header-only-bearer",
			authHeader:        "Bearer",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
		},
		{
			name:              "error-wrong-scheme",
			authHeader:        "Basic dXNlcjpwYXNz",
			expectedAbort:     true,
			expectedAbortCode: 401,
			expectedJWTSet:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			headers := make(http.Header)
			if test.authHeader != "" {
				headers.Set("Authorization", test.authHeader)
			}

			handler := service.JWTExtraction(headers)

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			handler(ctx)

			if test.expectedAbort {
				require.Equal(t, test.expectedAbortCode, w.Code)
			} else {
				require.Equal(t, http.StatusOK, w.Code)

				if test.expectedJWTSet {
					jwtValue, exists := ctx.Get("jwt")
					require.True(t, exists)
					require.NotNil(t, jwtValue)

					jwtArray, ok := jwtValue.([]string)
					require.True(t, ok)
					require.GreaterOrEqual(t, len(jwtArray), 2)
					require.Equal(t, "Bearer", jwtArray[0])
				}
			}
		})
	}
}
