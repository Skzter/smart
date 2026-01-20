package store

import (
	"sync"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
)

// TestLogStreamStore defines the contract for managing test log streams.
type TestLogStreamStore interface {
	// AddEvent adds a log event to the stream of a test. If the stream does not exist,
	// it is created automatically. If a completed stream already exists for this ID,
	// it is replaced by a new one to start a fresh log for a new execution.
	AddEvent(testId string, event entity.LogEvent)

	// GetStream retrieves the log stream for a given test ID and updates its last access time.
	// Returns the stream and a boolean indicating if it exists.
	GetStream(testId string) (*entity.LogStream, bool)

	// CompleteStream marks the log stream of a test as completed, preventing further events from being added.
	CompleteStream(testId string)

	// Shutdown stops the background cleanup routine and releases associated resources.
	Shutdown()
}

// TestLogStreamStore manages log streams for tests, providing thread-safe access
// and automatic cleanup of streams older than 24 hours.
type testLogStreamStore struct {
	streams         map[string]*entity.LogStream
	stopCleanUpCh   chan struct{}
	mu              sync.RWMutex
	cleanupInterval time.Duration
	streamMaxAge    time.Duration
}

// NewTestLogStreamStore creates a new TestLogStreamStore with default configuration and starts the cleanup routine.
// Default: cleanup runs every hour and removes streams older than 24 hours.
func NewTestLogStreamStore() TestLogStreamStore {
	return NewTestLogStreamStoreWithConfig(1*time.Hour, 24*time.Hour)
}

// NewTestLogStreamStoreWithConfig creates a new TestLogStreamStore with custom cleanup settings.
// This is primarily intended for testing purposes.
func NewTestLogStreamStoreWithConfig(cleanupInterval, streamMaxAge time.Duration) TestLogStreamStore {
	store := &testLogStreamStore{
		streams:         make(map[string]*entity.LogStream),
		stopCleanUpCh:   make(chan struct{}),
		cleanupInterval: cleanupInterval,
		streamMaxAge:    streamMaxAge,
	}
	go store.cleanup()
	return store
}

// AddEvent adds a log event to the stream of a test. If the stream does not exist,
// it is created automatically. If a completed stream already exists for this ID,
// it is replaced by a new one to ensure a clean state for subsequent test runs.
func (tlss *testLogStreamStore) AddEvent(testId string, event entity.LogEvent) {
	tlss.mu.Lock()
	stream, exists := tlss.streams[testId]
	if !exists || stream.IsCompleted() {
		stream = entity.NewLogStream()
		tlss.streams[testId] = stream
	}
	tlss.mu.Unlock()

	stream.AddEvent(event)
}

// GetStream retrieves the log stream for a given test ID and updates its last access timestamp
// to postpone automatic cleanup.
func (tlss *testLogStreamStore) GetStream(testId string) (*entity.LogStream, bool) {
	tlss.mu.RLock()
	defer tlss.mu.RUnlock()
	stream, exists := tlss.streams[testId]
	if exists {
		stream.UpdateLastAccess()
	}
	return stream, exists
}

// CompleteStream marks the log stream of a test as completed.
func (tlss *testLogStreamStore) CompleteStream(testId string) {
	if stream, exists := tlss.GetStream(testId); exists {
		stream.SetComplete()
	}
}

// Shutdown stops the background cleanup routine.
func (tlss *testLogStreamStore) Shutdown() {
	close(tlss.stopCleanUpCh)
}

// cleanup periodically removes log streams that are older than streamMaxAge and are marked as completed.
// It runs in a background goroutine and checks every cleanupInterval until Shutdown is called.
func (tlss *testLogStreamStore) cleanup() {
	ticker := time.NewTicker(tlss.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-tlss.stopCleanUpCh:
			return
		case <-ticker.C:
			tlss.mu.Lock()
			now := time.Now()
			for testId, stream := range tlss.streams {
				if now.Sub(stream.GetLastAccessedAt()) > tlss.streamMaxAge && stream.IsCompleted() {
					delete(tlss.streams, testId)
				}
			}
			tlss.mu.Unlock()
		}
	}
}
