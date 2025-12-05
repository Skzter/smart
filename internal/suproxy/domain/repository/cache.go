package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
)

// RedisClient wraps the underlying redis client for easier mocking.
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Close() error
}

// Cache defines the high-level cache abstraction used by the application.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

// redisCache implements Cache using a Redis backend.
type redisCache struct {
	log *slog.Logger
	rdb RedisClient
}

// NewRedisCache initializes a new Redis-based cache implementation.
func NewRedisCache(logger *slog.Logger, cfg *config.Config) (Cache, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, fmt.Errorf("logger cannot be nil, %w", err)
	}

	if err := assert.NotNil(cfg); err != nil {
		return nil, fmt.Errorf("config cannot be nil, %w", err)
	}

	// Default options
	opts := redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	}

	if cfg.Redis != nil {
		opts.Addr = cfg.Redis.Addr
		opts.Password = cfg.Redis.Password
		opts.DB = cfg.Redis.Db
		opts.Protocol = cfg.Redis.Protocol
	}

	client := redis.NewClient(&opts)

	return &redisCache{
		log: logger,
		rdb: client,
	}, nil
}

// Get retrieves a cached value by key and reports whether it was a hit.
func (r *redisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, false, fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return nil, false, fmt.Errorf("key cannot be empty, %w", err)
	}

	val, err := r.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		r.log.Error("redis: get failed", "key", key, "err", err)
		return nil, false, err
	}
	return val, true, nil
}

// Set stores a value with an expiration time.
func (r *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key cannot be empty, %w", err)
	}
	if err := assert.NotNil(value); err != nil {
		return fmt.Errorf("value cannot be nil, %w", err)
	}
	if err := r.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		r.log.Error("redis: set failed", "key", key, "ttl", ttl, "err", err)
		return err
	}

	r.log.Debug("redis: set", "key", key, "ttl", ttl)
	return nil
}

// Delete removes a value from the cache.
func (r *redisCache) Delete(ctx context.Context, key string) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key cannot be empty, %w", err)
	}
	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		return err
	}

	r.log.Debug("redis: delete", "key", key)
	return nil
}

// Close closes the Redis client connection.
func (r *redisCache) Close() error {
	return r.rdb.Close()
}
