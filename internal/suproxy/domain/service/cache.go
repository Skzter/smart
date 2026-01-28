package service

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// CacheService is an interface for implementing the cache infrastructure
type CacheService interface {
	Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error)
	Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error
	Invalidate(ctx context.Context, req entity.Request, isMock bool) error
	BuildKey(req entity.Request, isMock bool) string
}
