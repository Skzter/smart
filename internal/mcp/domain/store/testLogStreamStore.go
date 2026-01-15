package store

import (
	"sync"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
)

// TestLogStreamStore defines the contract for managing test log streams.
type TestLogStreamStore interface {
	// AddEvent adds a log event to the stream of a test. If the stream does not exist, it is created automatically.
	AddEvent(testId string, event entity.LogEvent)

	// GetStream retrieves the log stream for a given test ID.
	// Returns the stream and a boolean indicating if it exists.
	GetStream(testId string) (*entity.LogStream, bool)

	// CompleteStream marks the log stream of a test as completed.
	CompleteStream(testId string)
}

// TestLogStreamStore manages log streams for tests, providing thread-safe access
// and automatic cleanup of streams older than 24 hours.
type testLogStreamStore struct {
	streams map[string]*entity.LogStream
	mu      sync.RWMutex
}

// NewTestLogStreamStore creates a new TestLogStreamStore and starts the cleanup routine.
func NewTestLogStreamStore() TestLogStreamStore {
	store := &testLogStreamStore{
		streams: make(map[string]*entity.LogStream),
	}
	go store.cleanup()
	return store
}

// AddEvent adds a log event to the stream of a test. If the stream does not exist,
// it is created automatically.
func (tlss *testLogStreamStore) AddEvent(testId string, event entity.LogEvent) {
	tlss.mu.Lock()
	stream, exists := tlss.streams[testId]
	if !exists {
		stream = entity.NewLogStream()
		tlss.streams[testId] = stream
	}
	tlss.mu.Unlock()

	stream.AddEvent(event)
}

// GetStream retrieves the log stream for a given test ID.
// Returns the stream and a boolean indicating if it exists.
func (tlss *testLogStreamStore) GetStream(testId string) (*entity.LogStream, bool) {
	tlss.mu.RLock()
	defer tlss.mu.RUnlock()
	stream, exists := tlss.streams[testId]
	return stream, exists
}

// CompleteStream marks the log stream of a test as completed.
func (tlss *testLogStreamStore) CompleteStream(testId string) {
	if stream, exists := tlss.GetStream(testId); exists {
		stream.SetComplete()
	}
}

// cleanup periodically removes log streams that are older than 24 hours.
// It should run in a background goroutine and checks every hour.
func (tlss *testLogStreamStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tlss.mu.Lock()
		now := time.Now()
		for testId, stream := range tlss.streams {
			if now.Sub(stream.CreatedAt) > 24*time.Hour {
				delete(tlss.streams, testId)
			}
		}
		tlss.mu.Unlock()
	}
}
