package service

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// DatabaseService defines the business capability for interacting with stored entries.
type DatabaseService interface {
	SaveDbEntry(ctx context.Context, entry entity.DatabaseEntry) error
	GetAllKeys(ctx context.Context) ([]string, error)
}
