package store

import (
	"fmt"
	"testing"
	"time"

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			if test.setup != nil {
				test.setup(store, test.testId)
			}

			store.AddEvent(test.testId, test.event)

			test.verify(t, store, test.testId)
		})
	}
}

func TestGetStream(t *testing.T) {
	tests := []struct {
		name          string
		testId        string
		setup         func(store TestLogStreamStore, testId string)
		expectExists  bool
		expectEvents  int
		verifyDetails func(t *testing.T, stream *entity.LogStream)
	}{
		{
			name:         "get non-existent stream",
			testId:       "non-existent",
			setup:        nil,
			expectExists: false,
			expectEvents: 0,
		},
		{
			name:   "get existing stream",
			testId: "test-existing",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 1"})
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 2"})
			},
			expectExists: true,
			expectEvents: 2,
		},
		{
			name:   "get completed stream",
			testId: "test-completed",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})
				store.CompleteStream(testId)
			},
			expectExists: true,
			expectEvents: 1,
			verifyDetails: func(t *testing.T, stream *entity.LogStream) {
				assert.True(t, stream.IsCompleted(), "stream should be completed")
			},
		},
		{
			name:   "get stream updates last access time",
			testId: "test-access-time",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})
				time.Sleep(10 * time.Millisecond)
			},
			expectExists: true,
			expectEvents: 1,
			verifyDetails: func(t *testing.T, stream *entity.LogStream) {
				lastAccess := stream.GetLastAccessedAt()
				assert.WithinDuration(t, time.Now(), lastAccess, 100*time.Millisecond,
					"last access time should be updated to current time")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			if test.setup != nil {
				test.setup(store, test.testId)
			}

			stream, exists := store.GetStream(test.testId)

			assert.Equal(t, test.expectExists, exists, "existence should match expectation")

			if test.expectExists {
				require.NotNil(t, stream, "stream should not be nil when it exists")
				events := stream.GetEvents()
				assert.Len(t, events, test.expectEvents, "event count should match")

				if test.verifyDetails != nil {
					test.verifyDetails(t, stream)
				}
			}
		})
	}
}

func TestCompleteStream(t *testing.T) {
	tests := []struct {
		name   string
		testId string
		setup  func(store TestLogStreamStore, testId string)
		verify func(t *testing.T, store TestLogStreamStore, testId string)
	}{
		{
			name:   "complete existing stream",
			testId: "test-complete",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")
				assert.True(t, stream.IsCompleted(), "stream should be completed")
			},
		},
		{
			name:   "complete non-existent stream does not panic",
			testId: "non-existent",
			setup:  nil,
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				_, exists := store.GetStream(testId)
				assert.False(t, exists, "stream should not exist")
			},
		},
		{
			name:   "complete already completed stream",
			testId: "test-double-complete",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})
				store.CompleteStream(testId)
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")
				assert.True(t, stream.IsCompleted(), "stream should still be completed")
			},
		},
		{
			name:   "events cannot be added after completion",
			testId: "test-no-events-after-complete",
			setup: func(store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 1"})
				store.CompleteStream(testId)
			},
			verify: func(t *testing.T, store TestLogStreamStore, testId string) {
				store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event 2"})

				stream, exists := store.GetStream(testId)
				require.True(t, exists, "stream should exist")

				events := stream.GetEvents()
				assert.Len(t, events, 1, "new stream should have only new event after old was completed")
				assert.Equal(t, "Event 2", events[0].Data, "should be the new event")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			if test.setup != nil {
				test.setup(store, test.testId)
			}

			store.CompleteStream(test.testId)

			test.verify(t, store, test.testId)
		})
	}
}

func TestCleanup(t *testing.T) {
	tests := []struct {
		name          string
		cleanupAfter  time.Duration
		maxAge        time.Duration
		setup         func(store TestLogStreamStore)
		waitDuration  time.Duration
		verifyStreams func(t *testing.T, store TestLogStreamStore)
	}{
		{
			name:         "old completed streams are removed",
			cleanupAfter: 100 * time.Millisecond,
			maxAge:       50 * time.Millisecond,
			setup: func(store TestLogStreamStore) {
				store.AddEvent("old-completed", entity.LogEvent{Event: "log", Data: "Event"})
				store.CompleteStream("old-completed")
			},
			waitDuration: 200 * time.Millisecond,
			verifyStreams: func(t *testing.T, store TestLogStreamStore) {
				_, exists := store.GetStream("old-completed")
				assert.False(t, exists, "old completed stream should be removed")
			},
		},
		{
			name:         "recent streams are not removed",
			cleanupAfter: 100 * time.Millisecond,
			maxAge:       500 * time.Millisecond,
			setup: func(store TestLogStreamStore) {
				store.AddEvent("recent-1", entity.LogEvent{Event: "log", Data: "Event"})
				store.AddEvent("recent-2", entity.LogEvent{Event: "log", Data: "Event"})
				store.CompleteStream("recent-1")
			},
			waitDuration: 200 * time.Millisecond,
			verifyStreams: func(t *testing.T, store TestLogStreamStore) {
				_, exists1 := store.GetStream("recent-1")
				_, exists2 := store.GetStream("recent-2")
				assert.True(t, exists1, "recent completed stream should not be removed")
				assert.True(t, exists2, "recent active stream should not be removed")
			},
		},
		{
			name:         "incomplete streams are not removed even if old",
			cleanupAfter: 100 * time.Millisecond,
			maxAge:       50 * time.Millisecond,
			setup: func(store TestLogStreamStore) {
				store.AddEvent("incomplete", entity.LogEvent{Event: "log", Data: "Event"})
			},
			waitDuration: 200 * time.Millisecond,
			verifyStreams: func(t *testing.T, store TestLogStreamStore) {
				_, exists := store.GetStream("incomplete")
				assert.True(t, exists, "incomplete streams should never be removed")
			},
		},
		{
			name:         "multiple old completed streams are cleaned up",
			cleanupAfter: 100 * time.Millisecond,
			maxAge:       50 * time.Millisecond,
			setup: func(store TestLogStreamStore) {
				for i := 0; i < 5; i++ {
					testId := fmt.Sprintf("old-%d", i)
					store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})
					store.CompleteStream(testId)
				}
			},
			waitDuration: 200 * time.Millisecond,
			verifyStreams: func(t *testing.T, store TestLogStreamStore) {
				for i := range 5 {
					testId := fmt.Sprintf("old-%d", i)
					_, exists := store.GetStream(testId)
					assert.False(t, exists, "old stream %s should be removed", testId)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTestLogStreamStoreWithConfig(test.cleanupAfter, test.maxAge)
			defer store.Shutdown()

			test.setup(store)
			time.Sleep(test.waitDuration)
			test.verifyStreams(t, store)
		})
	}
}
