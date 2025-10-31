package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// CacheRepository abstracts the underlying cache storage
type CacheRepository interface {
	Get(ctx context.Context, key string) (value []byte, hit bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// CacheService defines caching operations for supplier and mock responses
type CacheService interface {
	Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error)
	Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error
	Invalidate(ctx context.Context, req entity.Request, isMock bool) error
	BuildKey(req entity.Request, isMock bool) string
}

type cacheService struct {
	log   *slog.Logger
	cfg   *config.Config
	repo  CacheRepository
	ttl   time.Duration
	nowFn func() time.Time
}

// CacheEntry represents the persisted payload
type CacheEntry struct {
	Response json.RawMessage `json:"response"`
	CachedAt time.Time       `json:"cached_at"`
}

// NewCacheService creates a v0.1 cache service with a static TTL
func NewCacheService(log *slog.Logger, cfg *config.Config, repo CacheRepository) CacheService {
	return &cacheService{
		log:   log,
		cfg:   cfg,
		repo:  repo,
		ttl:   5 * time.Minute,
		nowFn: time.Now,
	}
}

// Lookup retrieves a cached response if present; corrupt entries are treated as miss
func (s *cacheService) Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error) {
	key := s.BuildKey(req, isMock)

	raw, hit, err := s.repo.Get(ctx, key)
	if err != nil || !hit {
		return nil, false, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		s.log.Warn("cache: unmarshal failed, treating as miss", "key", key, "err", err)
		return nil, false, nil
	}

	return []byte(entry.Response), true, nil
}

// Store writes the response with a static TTL; ignores isError/isMock
func (s *cacheService) Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error {
	key := s.BuildKey(req, isMock)

	entry := CacheEntry{
		Response: json.RawMessage(response),
		CachedAt: s.nowFn().UTC(),
	}

	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return s.repo.Set(ctx, key, payload, s.ttl)
}

// Invalidate removes an entry; v0.1 is a thin wrapper.
func (s *cacheService) Invalidate(ctx context.Context, req entity.Request, isMock bool) error {
	key := s.BuildKey(req, isMock)
	return s.repo.Delete(ctx, key)
}

// BuildKey creates a naive key: prefix + destination + hashless body slice.
func (s *cacheService) BuildKey(req entity.Request, isMock bool) string {
	prefix := "suproxy:cache:v0.1"
	dest := strings.TrimSpace(req.Destination)

	// cheap body normalization
	body := strings.TrimSpace(req.Request)
	if len(body) > 128 {
		body = body[:128]
	}

	mode := "live"
	if isMock {
		mode = "mock"
	}

	// final key format (may contain spaces if dest/body include them)
	return strings.Join([]string{prefix, mode, dest, body}, "|")
}
