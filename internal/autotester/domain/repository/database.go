package repository

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
)

// TokenDatabase is for CRUD methods on db
type TokenDatabase interface {
	CreateToken(ctx context.Context, arg database.CreateTokenParams) error
	ReadToken(ctx context.Context, userID string) (database.RefreshToken, error)
	UpdateToken(ctx context.Context, arg database.UpdateTokenParams) error
}
