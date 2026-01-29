package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the bearer token before entering protected endpoints.
//
// Behavior:
// - 401 Unauthorized: missing/invalid Authorization header OR invalid/expired/unknown token OR validation error
// - 403 Forbidden: token is revoked
// - calls c.Next() if token is valid and not revoked
func (a *AutotesterController) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := a.authService.GetBearerToken(c.Request.Header)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := a.authService.ValidateToken(c.Request.Context(), token)
		if err != nil || res == nil || !res.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if res.Revoked {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
