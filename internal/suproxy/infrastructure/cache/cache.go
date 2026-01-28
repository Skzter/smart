package service

// #nosec G501 - MD5 is acceptable here for non-cryptographic cache key generation
import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"sort"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	sharedRepository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// CacheService interface defines the business logic for caching supplier or mock responses
type CacheService interface {
	Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error)
	Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error
	Invalidate(ctx context.Context, req entity.Request, isMock bool) error
	BuildKey(req entity.Request, isMock bool) string
}

// cacheService struct implements CacheService and contains the main cache logic
type cacheService struct {
	log    *slog.Logger
	cfg    *config.Suproxy
	repo   sharedRepository.Cache
	ttls   entity.CacheTTLPolicy
	tracer trace.Tracer
}

// NewCacheService creates and configures a new instance of cacheService with default TTLs
func NewCacheService(
	log *slog.Logger,
	cfg *config.Suproxy,
	repo sharedRepository.Cache,
	tracer trace.Tracer,
) (CacheService, error) {
	if err := assert.NotNil(log, cfg, repo, tracer); err != nil {
		return nil, err
	}

	ttls := entity.CacheTTLPolicy{
		// Default TTL configuration
		SupplierOK:   10 * time.Minute,
		MockOK:       20 * time.Minute,
		ErrorOrEmpty: 45 * time.Second,
	}

	svc := &cacheService{
		log:    log,
		cfg:    cfg,
		repo:   repo,
		ttls:   ttls,
		tracer: tracer,
	}

	log.Debug("cache: service initialized",
		"supplier_ttl", svc.ttls.SupplierOK,
		"mock_ttl", svc.ttls.MockOK,
		"error_or_empty_ttl", svc.ttls.ErrorOrEmpty,
	)

	return svc, nil
}

// Lookup checks if a cached entry exists for a given request and returns it if found
func (s *cacheService) Lookup(ctx context.Context, req entity.Request, isMock bool) ([]byte, bool, error) {
	if err := assert.NotNil(ctx, req); err != nil {
		return nil, false, err
	}

	ctx, span := s.tracer.Start(ctx, "cache.Lookup")
	defer span.End()

	key := s.BuildKey(req, isMock)

	raw, hit, err := s.repo.Get(ctx, key) // Try to get entry from cache

	if err != nil {
		span.RecordError(err)
		s.log.Error("cache: lookup failed, treating as miss", "err", err)
		return nil, false, nil
	}

	if !hit {
		return nil, false, nil
	}

	span.AddEvent(
		"cache.Lookup.hit",
		trace.WithAttributes(
			attribute.String("key", key),
			attribute.Bool("mock", isMock),
		),
	)

	var entry entity.CacheEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		span.RecordError(err)
		// Defensive: ignore corrupted cache entries
		s.log.Warn("cache: failed to unmarshal entry, ignoring", "err", err)
		return nil, false, nil
	}

	span.SetStatus(codes.Ok, "")

	return []byte(entry.Response), true, nil
}

// Store saves a new cache entry for a given request and response with a calculated TTL
func (s *cacheService) Store(ctx context.Context, req entity.Request, response []byte, isMock bool, isError bool) error {
	if err := assert.NotNil(ctx, req, response); err != nil {
		return err
	}

	ctx, span := s.tracer.Start(ctx, "cache.Store")
	defer span.End()

	key := s.BuildKey(req, isMock)

	// Build cache entry object with metadata
	entry := entity.CacheEntry{
		Mock:     isMock,
		Key:      key,
		Request:  req,
		CachedAt: time.Now().UTC(),
		Version:  1,
	}

	if len(response) > 0 {
		entry.Response = json.RawMessage(response)
	}

	payload, err := json.Marshal(entry) // Serialize to JSON
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal payload")
		s.log.Error("cache: failed to marshal entry", "err", err)
		return err
	}

	ttl := s.chooseTTL(isMock, isError, response) // Pick appropriate TTL

	span.SetAttributes(
		attribute.String("key", key),
		attribute.Bool("mock", isMock),
		attribute.Int64("ttl", int64(ttl)),
	)

	if err = s.repo.Set(ctx, key, payload, ttl); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to store entry")
		s.log.Error("cache: store failed", "err", err)
		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
}

// Invalidate removes a specific cache entry based on the request key
func (s *cacheService) Invalidate(ctx context.Context, req entity.Request, isMock bool) error {
	if err := assert.NotNil(ctx, req); err != nil {
		return err
	}

	ctx, span := s.tracer.Start(ctx, "cache.Invalidate")
	defer span.End()

	key := s.BuildKey(req, isMock) // Reconstruct cache key

	span.SetAttributes(
		attribute.String("key", key),
		attribute.Bool("mock", isMock),
	)

	if err := s.repo.Delete(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to invalidate entry")
		s.log.Error("cache: invalidation failed", "err", err)
		return err
	}

	span.SetStatus(codes.Ok, "")

	return nil
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
	b.WriteString(strings.TrimSpace(req.Body))
	return b.String()
}

// md5sum computes the MD5 hash of the input string and returns its hex encoding
func md5sum(s string) string {
	// #nosec G401 - MD5 used intentionally for fast, non-secure hashing
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// isLikelyEmptyJSON returns true if the given JSON looks empty or null-like
func isLikelyEmptyJSON(b []byte) bool {
	trim := strings.TrimSpace(string(b))
	return trim == "" || trim == "{}" || trim == "[]" || trim == "null"
}
