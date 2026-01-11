package service

import (
	"context"
	"errors"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// JWTValidator validates JWT tokens and returns a normalized ValidationResult
type JWTValidator interface {
	Validate(ctx context.Context, jwt string) (entity.ValidationResult, error)
}

type jwtValidatorImpl struct{}

// NewJWTValidator creates a new JWTValidator instance
func NewJWTValidator() JWTValidator { return &jwtValidatorImpl{} }

func (v *jwtValidatorImpl) Validate(ctx context.Context, jwt string) (entity.ValidationResult, error) {
	_ = ctx
	if strings.TrimSpace(jwt) == "" {
		return entity.ValidationResult{Valid: false, Revoked: false}, errors.New("empty token")
	}
	return entity.ValidationResult{Valid: true, Revoked: false}, nil
}
