package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogStream(t *testing.T) {
	stream := NewLogStream()

	assert.NotNil(t, stream)
	assert.False(t, stream.IsCompleted())
	assert.Empty(t, stream.GetEvents())
	assert.WithinDuration(t, time.Now(), stream.GetLastAccessedAt(), 100*time.Millisecond)
}

func TestAddEvent(t *testing.T) {
	tests := []struct {
		name           string
		events         []LogEvent
		completeAfter  int
		expectedEvents int
	}{
		{
			name: "add single event",
			events: []LogEvent{
				{Event: "log", Data: "First event"},
			},
			completeAfter:  -1,
			expectedEvents: 1,
		},
		{
			name: "add multiple events",
			events: []LogEvent{
				{Event: "log", Data: "Event 1"},
				{Event: "log", Data: "Event 2"},
				{Event: "log", Data: "Event 3"},
			},
			completeAfter:  -1,
			expectedEvents: 3,
		},
		{
			name: "cannot add event after completion",
			events: []LogEvent{
				{Event: "log", Data: "Event 1"},
				{Event: "log", Data: "Event 2"},
			},
			completeAfter:  0,
			expectedEvents: 0,
		},
		{
			name: "add events with different types",
			events: []LogEvent{
				{Event: "log", Data: "Log message"},
				{Event: "error", Data: "Error message"},
				{Event: "done", Data: "Done message"},
			},
			completeAfter:  -1,
			expectedEvents: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()

			for i, event := range test.events {
				if test.completeAfter == i {
					stream.SetComplete()
				}
				stream.AddEvent(event)
			}

			events := stream.GetEvents()
			assert.Len(t, events, test.expectedEvents)
		})
	}
}

func TestSetComplete(t *testing.T) {
	tests := []struct {
		name        string
		addEvents   int
		verifyState func(t *testing.T, stream *LogStream)
	}{
		{
			name:      "mark stream as completed",
			addEvents: 2,
			verifyState: func(t *testing.T, stream *LogStream) {
				assert.True(t, stream.IsCompleted())
			},
		},
		{
			name:      "complete empty stream",
			addEvents: 0,
			verifyState: func(t *testing.T, stream *LogStream) {
				assert.True(t, stream.IsCompleted())
				assert.Empty(t, stream.GetEvents())
			},
		},
		{
			name:      "complete multiple times is idempotent",
			addEvents: 1,
			verifyState: func(t *testing.T, stream *LogStream) {
				assert.True(t, stream.IsCompleted())
			},
		},
		{
			name:      "updates last access time on completion",
			addEvents: 1,
			verifyState: func(t *testing.T, stream *LogStream) {
				assert.WithinDuration(t, time.Now(), stream.GetLastAccessedAt(), 100*time.Millisecond)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()

			for i := 0; i < test.addEvents; i++ {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event"})
			}

			if test.name == "complete multiple times is idempotent" {
				stream.SetComplete()
			}
			stream.SetComplete()

			test.verifyState(t, stream)
		})
	}
}

func TestGetEvents(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(stream *LogStream)
		expectedEvents int
		expectNil      bool
	}{
		{
			name: "get events from active stream returns copy",
			setup: func(stream *LogStream) {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event 1"})
				stream.AddEvent(LogEvent{Event: "log", Data: "Event 2"})
			},
			expectedEvents: 2,
			expectNil:      false,
		},
		{
			name: "get events from completed stream returns direct reference",
			setup: func(stream *LogStream) {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event 1"})
				stream.SetComplete()
			},
			expectedEvents: 1,
			expectNil:      false,
		},
		{
			name:           "get events from empty stream returns nil",
			setup:          func(stream *LogStream) {},
			expectedEvents: 0,
			expectNil:      true,
		},
		{
			name: "get events updates last access time",
			setup: func(stream *LogStream) {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event"})
				time.Sleep(10 * time.Millisecond)
			},
			expectedEvents: 1,
			expectNil:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			initialTime := stream.GetLastAccessedAt()

			test.setup(stream)

			events := stream.GetEvents()

			if test.expectNil {
				assert.Nil(t, events)
			} else {
				require.NotNil(t, events)
				assert.Len(t, events, test.expectedEvents)

				if test.name == "get events updates last access time" {
					assert.True(t, stream.GetLastAccessedAt().After(initialTime))
				}
			}
		})
	}
}

func TestIsCompleted(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(stream *LogStream)
		expectedCompleted bool
	}{
		{
			name:              "new stream is not completed",
			setup:             func(stream *LogStream) {},
			expectedCompleted: false,
		},
		{
			name: "stream with events is not completed",
			setup: func(stream *LogStream) {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event"})
			},
			expectedCompleted: false,
		},
		{
			name: "completed stream returns true",
			setup: func(stream *LogStream) {
				stream.AddEvent(LogEvent{Event: "log", Data: "Event"})
				stream.SetComplete()
			},
			expectedCompleted: true,
		},
		{
			name: "empty completed stream returns true",
			setup: func(stream *LogStream) {
				stream.SetComplete()
			},
			expectedCompleted: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			test.setup(stream)

			completed := stream.IsCompleted()
			assert.Equal(t, test.expectedCompleted, completed)
		})
	}
}

func TestUpdateLastAccess(t *testing.T) {
	tests := []struct {
		name       string
		waitBefore time.Duration
	}{
		{
			name:       "updates last access timestamp",
			waitBefore: 10 * time.Millisecond,
		},
		{
			name:       "updates multiple times",
			waitBefore: 5 * time.Millisecond,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			initialTime := stream.GetLastAccessedAt()

			time.Sleep(test.waitBefore)
			stream.UpdateLastAccess()

			updatedTime := stream.GetLastAccessedAt()
			assert.True(t, updatedTime.After(initialTime))

			if test.name == "updates multiple times" {
				time.Sleep(test.waitBefore)
				stream.UpdateLastAccess()
				finalTime := stream.GetLastAccessedAt()
				assert.True(t, finalTime.After(updatedTime))
			}
		})
	}
}

func TestGetLastAccessedAt(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(stream *LogStream) time.Time
		verify func(t *testing.T, stream *LogStream, expectedTime time.Time)
	}{
		{
			name: "returns initial timestamp",
			setup: func(stream *LogStream) time.Time {
				return stream.GetLastAccessedAt()
			},
			verify: func(t *testing.T, stream *LogStream, expectedTime time.Time) {
				assert.Equal(t, expectedTime, stream.GetLastAccessedAt())
			},
		},
		{
			name: "returns updated timestamp after access",
			setup: func(stream *LogStream) time.Time {
				time.Sleep(10 * time.Millisecond)
				stream.UpdateLastAccess()
				return stream.GetLastAccessedAt()
			},
			verify: func(t *testing.T, stream *LogStream, expectedTime time.Time) {
				assert.WithinDuration(t, expectedTime, stream.GetLastAccessedAt(), time.Millisecond)
			},
		},
		{
			name: "returns updated timestamp after completion",
			setup: func(stream *LogStream) time.Time {
				time.Sleep(10 * time.Millisecond)
				stream.SetComplete()
				return stream.GetLastAccessedAt()
			},
			verify: func(t *testing.T, stream *LogStream, expectedTime time.Time) {
				assert.WithinDuration(t, expectedTime, stream.GetLastAccessedAt(), time.Millisecond)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			expectedTime := test.setup(stream)
			test.verify(t, stream, expectedTime)
		})
	}
}
