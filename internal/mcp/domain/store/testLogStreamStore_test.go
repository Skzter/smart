package store

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
)

func TestNewTestLogStreamStore(t *testing.T) {
	store := NewTestLogStreamStore()
	defer store.Shutdown()

	assert.NotNil(t, store)
}

// nolint:funlen
func TestAddEvent(t *testing.T) {
	tests := []struct {
		name   string
		testId string
		event  entity.LogEvent
		setup  func(store TestLogStreamStore, testId string)
		verify func(t *testing.T, store TestLogStreamStore, testId string)
	}{
		{
			name:   "add event to new stream",
			testId: "test-new",
			event:  entity.LogEvent{Event: "log", Data: "First event"},
			setup:  nil,
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")
				require.NotNil(t, stream, "stream should not be nil")

				events := stream.GetEvents()
				require.Len(t, events, 1, "stream should have exactly 1 event")
				assert.Equal(t, "log", events[0].Event)
				assert.Equal(t, "First event", events[0].Data)
			},
		},
		{
			name:   "add event to existing stream",
			testId: "test-existing",
			event:  entity.LogEvent{Event: "log", Data: "Second event"},
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "First event"})
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")

				events := stream.GetEvents()
				require.Len(t, events, 2, "stream should have exactly 2 events")
				assert.Equal(t, "First event", events[0].Data)
				assert.Equal(t, "Second event", events[1].Data)
			},
		},
		{
			name:   "add event to completed stream creates new stream",
			testId: "test-completed",
			event:  entity.LogEvent{Event: "log", Data: "New event after completion"},
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Old event"})
				store.CompleteStream(testId)
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")

				assert.False(t, stream.IsCompleted(), "new stream should not be completed")

				events := stream.GetEvents()
				require.Len(t, events, 1, "new stream should have only 1 event")
				assert.Equal(t, "New event after completion", events[0].Data)
			},
		},
		{
			name:   "add multiple events sequentially",
			testId: "test-multiple",
			event:  entity.LogEvent{Event: "log", Data: "Event 3"},
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 1"})
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 2"})
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")

				events := stream.GetEvents()
				require.Len(t, events, 3, "stream should have exactly 3 events")
				assert.Equal(t, "Event 1", events[0].Data)
				assert.Equal(t, "Event 2", events[1].Data)
				assert.Equal(t, "Event 3", events[2].Data)
			},
		},
		{
			name:   "add event with different event types",
			testId: "test-event-types",
			event:  entity.LogEvent{Event: "error", Data: "Error message"},
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Log message"})
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")

				events := stream.GetEvents()
				require.Len(t, events, 2)
				assert.Equal(t, "log", events[0].Event)
				assert.Equal(t, "error", events[1].Event)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			if tt.setup != nil {
				tt.setup(store, tt.testId)
			}

			store.AddEvent(tt.testId, tt.event)

			tt.verify(t, store, tt.testId)
		})
	}
}

func TestAddEvent_Concurrent(t *testing.T) {
	tests := []struct {
		name               string
		numGoroutines      int
		eventsPerGoroutine int
	}{
		{
			name:               "concurrent writes to same stream",
			numGoroutines:      50,
			eventsPerGoroutine: 20,
		},
		{
			name:               "high concurrency stress test",
			numGoroutines:      100,
			eventsPerGoroutine: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			testId := "concurrent-test"
			var wg sync.WaitGroup
			wg.Add(tt.numGoroutines)

			for i := 0; i < tt.numGoroutines; i++ {
				go func(goroutineId int) {
					defer wg.Done()
					for j := 0; j < tt.eventsPerGoroutine; j++ {
						event := entity.LogEvent{
							Event: "log",
							Data:  fmt.Sprintf("Goroutine %d - Event %d", goroutineId, j),
						}
						store.AddEvent(testId, event)
					}
				}(i)
			}

			wg.Wait()

			stream, exists := store.GetStream(testId)
			require.True(t, exists, "stream should exist")

			events := stream.GetEvents()
			expectedCount := tt.numGoroutines * tt.eventsPerGoroutine
			assert.Equal(t, expectedCount, len(events),
				"all events should be present without race conditions")
		})
	}
}
