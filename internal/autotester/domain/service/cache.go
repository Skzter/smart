package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Cache is the Interface which holds all Methods for the Autotester Cache Service
// This Cache Service is responsible for caching the chats from the users
type Cache interface {
	LookUp(ctx context.Context, chatId string) (*entity.Chat, error)
	StoreOrReplace(ctx context.Context, chat entity.Chat, key string, timeToLive time.Duration) error
	/* GenerateKey */
	/* Init */
}

type cache struct {
	config    *config.Autotester
	logger    *slog.Logger
	cacheRepo sharedRepo.Cache
	tracer    trace.Tracer
}

// NewCacheService creates a new CacheService with config, logger, cacheRepo and tracer
// with this service you can LookUp Chats via ChatId and StoreOrReplace Chats in the Cache with an Chat, Key and TimeToLive
func NewCacheService(config *config.Autotester, logger *slog.Logger, cacheRepo sharedRepo.Cache, tracer trace.Tracer) (Cache, error) {
	if err := assert.NotNil(config, logger, cacheRepo, tracer); err != nil {
		return nil, err
	}

	/* maybe init cache here with chats */

	return &cache{
		config:    config,
		logger:    logger,
		cacheRepo: cacheRepo,
		tracer:    tracer,
	}, nil
}

// LookUp searches the Cache for a chat with the given chatId
// if successful, it returns the found chat
// if a cache miss occurs, it will return a nil chat and no error
// if anything else fails, it will return a nil chat and an error
func (c *cache) LookUp(ctx context.Context, chatId string) (*entity.Chat, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	if err := assert.StringNotEmpty(chatId); err != nil {
		return nil, err
	}

	ctx, span := c.tracer.Start(ctx, "AutotesterCacheService.LookUp")
	defer span.End()

	entry, hit, err := c.cacheRepo.Get(ctx, chatId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache repo get error")
		return nil, err
	}
	if !hit {
		return nil, nil
	}
	var chat entity.Chat
	if err := json.Unmarshal(entry, &chat); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache entry unmarshal error")
		return nil, err
	}
	span.SetStatus(codes.Ok, "")
	return &chat, nil
}

// StoreOrReplace stores the Chat when the Cache is not full, if the Cache is full it will replace an entry with this new one
// CacheReplacementStrategy: tbd (maybe Least Recently Used)
// if successful, no error
// else, error
func (c *cache) StoreOrReplace(ctx context.Context, chat entity.Chat, key string, timeToLive time.Duration) error {
	return nil
}
