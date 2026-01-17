package entity

import (
	"sync"
	"time"
)

// LogStream represents a stream of log events with thread-safe access and creation timestamp.
type LogStream struct {
	events    []LogEvent
	completed bool
	CreatedAt time.Time
	mu        sync.RWMutex
}

// NewLogStream creates a new LogStream with an empty events list and current creation timestamp.
func NewLogStream() *LogStream {
	return &LogStream{
		events:    make([]LogEvent, 0, 5),
		completed: false,
		CreatedAt: time.Now(),
	}
}

// AddEvent appends a log event to the stream in a thread-safe manner.
func (ls *LogStream) AddEvent(event LogEvent) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.events = append(ls.events, event)
}

// SetComplete marks the log stream as completed in a thread-safe manner.
func (ls *LogStream) SetComplete() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.completed = true
}

// GetEvents returns a copy of all log events in the stream in a thread-safe manner.
func (ls *LogStream) GetEvents() []LogEvent {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	eventsCopy := make([]LogEvent, len(ls.events))
	copy(eventsCopy, ls.events)
	return eventsCopy
}

// IsCompleted checks whether the log stream has been marked as completed in a thread-safe manner.
func (ls *LogStream) IsCompleted() bool {
	ls.mu.RLock()
	defer ls.mu.RUnlock()
	return ls.completed
}
