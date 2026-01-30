package handler

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// JwtExtractionHandler provides JWT extractor
type JwtExtractionHandler interface {
	JWTExtractionIntoContext() gin.HandlerFunc
}

type jwtExtractionHandler struct {
	logger        *slog.Logger
	jwtContextKey entity.JwtContextKey
}

// NewJWTAuthentification creates a new service for JWTAuthentification
func NewJWTAuthentification(logger *slog.Logger) (JwtExtractionHandler, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}
	return &jwtExtractionHandler{
		logger:        logger,
		jwtContextKey: entity.JwtContextKey{},
	}, nil
}

// JWTExtractionIntoContext extracts the token and pass it to gin.Context
func (j *jwtExtractionHandler) JWTExtractionIntoContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "empty bearer token"})
			return
		}

		ctx := context.WithValue(c.Request.Context(), j.jwtContextKey, token)
		c.Request = c.Request.WithContext(ctx)
		j.logger.Debug("JWTExtraction - wrote jwt to context.")
		c.Next()
	}
}
