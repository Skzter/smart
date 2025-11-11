package repository

import (
	"context"
	"time"
)

// CacheRepository defines the interface for interacting with the cache storage
type CacheRepository interface {
	Get(ctx context.Context, key string) (value []byte, hit bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// dummyCacheRepository is a temporary stub implementation
type dummyCacheRepository struct{}

// NewDummyCacheRepository creates a placeholder repository implementation
// This allows the code to compile until the real Redis logic is implemented
func NewDummyCacheRepository() CacheRepository {
	return &dummyCacheRepository{}
}

// Get always returns a cache miss (no data)
func (r *dummyCacheRepository) Get(ctx context.Context, key string) ([]byte, bool, error) {
	return nil, false, nil
}

// Set does nothing (simulates a no-op write)
func (r *dummyCacheRepository) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete does nothing (simulates a successful delete)
func (r *dummyCacheRepository) Delete(ctx context.Context, key string) error {
	return nil
}
