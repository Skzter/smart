package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// RedisClient wraps the underlying redis client for easier mocking.
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, ttl time.Duration) *redis.StatusCmd
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
	log    *slog.Logger
	rdb    RedisClient
	tracer trace.Tracer
}

// NewRedisCache initializes a new Redis-based cache implementation.
func NewRedisCache(
	logger *slog.Logger,
	cfg *config.RedisConfig,
	tracer trace.Tracer,
) (Cache, error) {
	if err := assert.NotNil(logger, cfg, tracer); err != nil {
		return nil, err
	}

	opts := redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
		Protocol: cfg.Protocol,
	}

	client := redis.NewClient(&opts)

	return &redisCache{
		log:    logger,
		rdb:    client,
		tracer: tracer,
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

	ctx, span := r.tracer.Start(ctx, "redisCache.Get")
	defer span.End()

	val, err := r.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get data")
		r.log.Error("redis: get failed", "key", key, "err", err)
		return nil, false, err
	}

	span.SetStatus(codes.Ok, "")

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

	ctx, span := r.tracer.Start(ctx, "redisCache.Set")
	defer span.End()

	if err := r.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to set data")
		r.log.Error("redis: set failed", "key", key, "ttl", ttl, "err", err)
		return err
	}

	span.SetStatus(codes.Ok, "")

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

	ctx, span := r.tracer.Start(ctx, "redisCache.Delete")
	defer span.End()

	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete data")
		return err
	}

	r.log.Debug("redis: delete", "key", key)
	return nil
}

// Close closes the Redis client connection.
func (r *redisCache) Close() error {
	return r.rdb.Close()
}
