package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

func TestAuthMiddleware_NoAuthHeader_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := mockService.NewMockAuth(t)

	ctrl := &AutotesterController{
		authService: mockAuth,
	}

	r := gin.New()
	r.Use(ctrl.AuthMiddleware())
	r.GET("/protected", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	mockAuth.
		On("GetBearerToken", mock.Anything).
		Return("", errors.New("missing auth header")).
		Once()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
