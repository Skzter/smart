package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// JWTValidator validates JWT tokens and returns a normalized ValidationResult
type JWTValidator interface {
	Validate(ctx context.Context, jwtString string) (entity.ValidationResult, error)
}

type jwtValidatorImpl struct {
	secret []byte
}

// NewJWTValidator creates a new JWTValidator instance
// Reads secret from env JWT_SECRET
func NewJWTValidator() (JWTValidator, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return &jwtValidatorImpl{secret: []byte(secret)}, nil
}

func (v *jwtValidatorImpl) Validate(ctx context.Context, jwtString string) (entity.ValidationResult, error) {
	_ = ctx

	jwtString = strings.TrimSpace(jwtString)
	if jwtString == "" {
		return entity.ValidationResult{Valid: false, Revoked: false}, errors.New("empty token")
	}

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(jwtString, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return entity.ValidationResult{Valid: false, Revoked: false}, fmt.Errorf("invalid token: %w", err)
	}

	// Validate exp at least
	if claims.ExpiresAt == nil {
		return entity.ValidationResult{Valid: false, Revoked: false}, errors.New("missing exp claim")
	}
	if time.Now().After(claims.ExpiresAt.Time) {
		return entity.ValidationResult{Valid: false, Revoked: false}, errors.New("token expired")
	}

	return entity.ValidationResult{Valid: true, Revoked: false}, nil
}
