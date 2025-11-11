package repository

import (
	"context"
	"sync"
	"time"
)

// CacheRepository defines the interface for interacting with the cache storage
type CacheRepository interface {
	Get(ctx context.Context, key string) (value []byte, hit bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// memoryCache ist eine einfache Thread-safe Map-basierte Cache-Implementierung.
type memoryCache struct {
	mu    sync.RWMutex
	store map[string]cacheEntry
}

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

// NewMemoryCacheRepository erstellt eine neue In-Memory-Cache Instanz.
func NewMemoryCacheRepository() CacheRepository {
	return &memoryCache{
		store: make(map[string]cacheEntry),
	}
}

// Get ruft einen Wert aus dem Cache ab, wenn er existiert und nicht abgelaufen ist.
func (m *memoryCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.store[key]
	if !ok {
		return nil, false, nil
	}

	if time.Now().After(entry.expiresAt) {
		// Eintrag ist abgelaufen – löschen und als Miss behandeln
		go func() {
			m.mu.Lock()
			delete(m.store, key)
			m.mu.Unlock()
		}()
		return nil, false, nil
	}

	return entry.value, true, nil
}

// Set speichert einen Wert mit TTL im Cache.
func (m *memoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[key] = cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Delete entfernt einen Eintrag aus dem Cache.
func (m *memoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, key)
	return nil
}
