package repository

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/infrastructure/database"
)

// TokenDatabase is for CRUD methods on db
type TokenDatabase interface {
	ReadToken(ctx context.Context, userID string) (database.RefreshToken, error)
	UpsertToken(ctx context.Context, arg database.UpsertTokenParams) (database.RefreshToken, error)
	ReadTokenByToken(ctx context.Context, token string) (database.RefreshToken, error)
}
