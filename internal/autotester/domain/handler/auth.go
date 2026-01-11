package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HandleValidateJWT validates the Authorization header JWT
func (c *AutotesterController) HandleValidateJWT(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	result, err := c.jwtValidator.Validate(ctx, token)
	if err != nil || !result.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if result.Revoked {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	ctx.Status(http.StatusOK)
}
