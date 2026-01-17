package entity

import (
	"sync"
	"time"
)

// LogStream represents a stream of log events with thread-safe access and creation timestamp.
type LogStream struct {
	events         []LogEvent
	completed      bool
	createdAt      time.Time
	lastAccessedAt time.Time
	mu             sync.RWMutex
}

// NewLogStream creates a new LogStream with an empty events list and current creation timestamp.
func NewLogStream() *LogStream {
	now := time.Now()
	return &LogStream{
		events:         make([]LogEvent, 0, 5),
		completed:      false,
		createdAt:      now,
		lastAccessedAt: now,
	}
}

// AddEvent appends a log event to the stream in a thread-safe manner.
// If the stream is already marked as completed, the event will be ignored to ensure
// the integrity of the finished log.
func (ls *LogStream) AddEvent(event LogEvent) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.completed {
		return
	}
	ls.events = append(ls.events, event)
}

// SetComplete marks the log stream as completed and updates the last access timestamp.
// Once completed, no more events can be added, and GetEvents will return direct references.
func (ls *LogStream) SetComplete() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.completed = true
	ls.lastAccessedAt = time.Now()
}

// GetEvents returns the log events in a thread-safe manner.
// If the stream is still active, it returns a copy to prevent data races with concurrent writers.
// Once the stream is completed, it returns a direct reference to the events slice to avoid
// expensive allocations and copying.
func (ls *LogStream) GetEvents() []LogEvent {
	ls.UpdateLastAccess()
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	if len(ls.events) == 0 {
		return nil
	}

	if !ls.completed {
		eventsCopy := make([]LogEvent, len(ls.events))
		copy(eventsCopy, ls.events)
		return eventsCopy
	}

	return ls.events
}

// IsCompleted checks whether the log stream has been marked as completed in a thread-safe manner.
func (ls *LogStream) IsCompleted() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.completed
}

// UpdateLastAccess updates the last access timestamp to the current time.
// This is used to track activity for the automatic cleanup mechanism.
func (ls *LogStream) UpdateLastAccess() {
	ls.mu.Lock()
	ls.lastAccessedAt = time.Now()
	ls.mu.Unlock()
}

// GetLastAccessedAt returns the timestamp of the last access or modification in a thread-safe manner.
func (ls *LogStream) GetLastAccessedAt() time.Time {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.lastAccessedAt
}
