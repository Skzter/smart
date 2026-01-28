package service

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// JwtExtractionService provides JWT extractor
type JwtExtractionService interface {
	JWTExtraction() gin.HandlerFunc
}

type jwtExtractionService struct {
	logger *slog.Logger
}

// NewJWTAuthentification creates a new service for JWTAuthentification
func NewJWTAuthentification(logger *slog.Logger) (JwtExtractionService, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}
	return &jwtExtractionService{
		logger: logger,
	}, nil
}

// JWTExtraction extracts the token and pass it to gin.Context
func (j *jwtExtractionService) JWTExtraction() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "token is missing"})
			return
		}
		splitHeader := strings.Split(header, " ")
		if len(splitHeader) < 2 || splitHeader[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "malformed authorization header"})
			return
		}

		c.Set("jwt", splitHeader[1])
		j.logger.Info("jwt: " + splitHeader[1])
		c.Next()
	}
}
