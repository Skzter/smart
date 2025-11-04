package service

import (
	"context"
	// #nosec G501 -- MD5 is acceptable here for non-cryptographic cache key generation
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"sort"
	"strings"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// CacheRepository defines the interface for interacting with the cache storage (Redis)
type CacheRepository interface {
	Get(ctx context.Context, key string) (value []byte, hit bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// CacheService interface defines the business logic for caching supplier or mock responses
type CacheService interface {
	Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error)
	Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error
	Invalidate(ctx context.Context, req entity.Request, isMock bool) error
	BuildKey(req entity.Request, isMock bool) string
}

// cacheService struct implements CacheService and contains the main cache logic
type cacheService struct {
	log   *slog.Logger
	cfg   *config.Config
	repo  CacheRepository
	ttls  ttlPolicy
	nowFn func() time.Time
}

// ttlPolicy defines TTL durations for different cache cases
type ttlPolicy struct {
	SupplierOK   time.Duration
	MockOK       time.Duration
	ErrorOrEmpty time.Duration
}

// CacheEntry represents how a single cache item is stored in Redis
type CacheEntry struct {
	Mock     bool            `json:"mock"`
	Key      string          `json:"key"`
	Request  entity.Request  `json:"request"`
	Response json.RawMessage `json:"response"`
	CachedAt time.Time       `json:"cached_at"`
	Version  int             `json:"v"`
}

// NewCacheService creates and configures a new instance of cacheService with default TTLs
func NewCacheService(log *slog.Logger, cfg *config.Config, repo CacheRepository) CacheService {
	ttls := ttlPolicy{
		// Default TTL configuration
		SupplierOK:   10 * time.Minute,
		MockOK:       20 * time.Minute,
		ErrorOrEmpty: 45 * time.Second,
	}

	// Return initialized cache service
	return &cacheService{
		log:   log,
		cfg:   cfg,
		repo:  repo,
		ttls:  ttls,
		nowFn: time.Now,
	}
}

// Lookup checks if a cached entry exists for a given request and returns it if found
func (s *cacheService) Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error) {
	key := s.BuildKey(req, isMock)

	raw, hit, err := s.repo.Get(ctx, key) // Try to get entry from cache
	if err != nil || !hit {
		return nil, false, err // Return miss or error
	}

	var entry CacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		// Defensive: ignore corrupted cache entries
		s.log.Warn("cache: failed to unmarshal entry, ignoring", "key", key, "err", err)
		return nil, false, nil
	}
	return []byte(entry.Response), true, nil
}

// Store saves a new cache entry for a given request and response with a calculated TTL
func (s *cacheService) Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error {
	key := s.BuildKey(req, isMock)

	// Build cache entry object with metadata
	entry := CacheEntry{
		Mock:     isMock,
		Key:      key,
		Request:  req,
		Response: json.RawMessage(response),
		CachedAt: s.nowFn().UTC(),
		Version:  1,
	}

	payload, err := json.Marshal(entry) // Serialize to JSON
	if err != nil {
		return err
	}

	ttl := s.chooseTTL(isMock, isError, response) // Pick appropriate TTL

	return s.repo.Set(ctx, key, payload, ttl) // Store in cache repository
}

// Invalidate removes a specific cache entry based on the request key
func (s *cacheService) Invalidate(ctx context.Context, req entity.Request, isMock bool) error {
	key := s.BuildKey(req, isMock) // Reconstruct cache key
	return s.repo.Delete(ctx, key) // Delete entry from cache
}

// BuildKey generates a stable cache key from request details (destination, headers, body)
func (s *cacheService) BuildKey(req entity.Request, isMock bool) string {
	prefix := "suproxy" // Static prefix for all keys

	mode := "live" // Default mode is supplier flow
	if isMock {
		mode = "mock" // Use "mock" if entry is from mock response
	}

	hash := md5sum(canonicalizeRequest(req))                        // Hash canonical request
	return strings.Join([]string{prefix, "cache", mode, hash}, ":") // Final key format
}

// chooseTTL decides the TTL based on response type and error state
func (s *cacheService) chooseTTL(isMock bool, isError bool, resp []byte) time.Duration {
	if isError || len(resp) == 0 || isLikelyEmptyJSON(resp) {
		return s.ttls.ErrorOrEmpty
	}
	if isMock {
		return s.ttls.MockOK // Longer TTL for mock responses
	}
	return s.ttls.SupplierOK // Default TTL for supplier responses
}

// canonicalizeRequest normalizes the request (dest + headers + body) to ensure stable hashing
func canonicalizeRequest(req entity.Request) string {
	var b strings.Builder

	// Add destination (trimmed) to canonical form
	b.WriteString(strings.TrimSpace(req.Destination))
	b.WriteString("\n")

	// Sort headers alphabetically for deterministic output
	keys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Append headers as "key:value" lines
	for _, k := range keys {
		v := req.Header[k]
		b.WriteString(strings.ToLower(strings.TrimSpace(k)))
		b.WriteString(":")
		b.WriteString(strings.TrimSpace(v))
		b.WriteString("\n")
	}

	// Append body content (trimmed)
	b.WriteString(strings.TrimSpace(req.Request))
	return b.String()
}

// md5sum computes the MD5 hash of the input string and returns its hex encoding
func md5sum(s string) string {
	// #nosec G401 -- MD5 used intentionally for fast, non-secure hashing
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// isLikelyEmptyJSON returns true if the given JSON looks empty or null-like
func isLikelyEmptyJSON(b []byte) bool {
	trim := strings.TrimSpace(string(b))
	return trim == "" || trim == "{}" || trim == "[]" || trim == "null"
}
