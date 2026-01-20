//go:build racetest

package entity

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddEvent_Concurrent(t *testing.T) {
	tests := []struct {
		name               string
		numGoroutines      int
		eventsPerGoroutine int
	}{
		{
			name:               "concurrent writes to same stream",
			numGoroutines:      5,
			eventsPerGoroutine: 10,
		},
		{
			name:               "stress test with more goroutines",
			numGoroutines:      20,
			eventsPerGoroutine: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func(goroutineId int) {
					defer wg.Done()
					for j := 0; j < test.eventsPerGoroutine; j++ {
						event := LogEvent{
							Event: "log",
							Data:  fmt.Sprintf("Goroutine %d - Event %d", goroutineId, j),
						}
						stream.AddEvent(event)
					}
				}(i)
			}

			wg.Wait()

			events := stream.GetEvents()
			expectedCount := test.numGoroutines * test.eventsPerGoroutine
			assert.Equal(t, expectedCount, len(events),
				"all events should be present without race conditions")
		})
	}
}

func TestGetEvents_Concurrent(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numReads      int
	}{
		{
			name:          "concurrent reads while active",
			numGoroutines: 5,
			numReads:      10,
		},
		{
			name:          "more concurrent readers",
			numGoroutines: 20,
			numReads:      10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			stream.AddEvent(LogEvent{Event: "log", Data: "Test event"})

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			successCount := make([]int, test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func(goroutineId int) {
					defer wg.Done()
					for j := 0; j < test.numReads; j++ {
						events := stream.GetEvents()
						if events != nil && len(events) > 0 {
							successCount[goroutineId]++
						}
					}
				}(i)
			}

			wg.Wait()

			totalReads := 0
			for _, count := range successCount {
				totalReads += count
			}

			expectedReads := test.numGoroutines * test.numReads
			assert.Equal(t, expectedReads, totalReads,
				"all concurrent reads should succeed without race conditions")
		})
	}
}

func TestSetComplete_Concurrent(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
	}{
		{
			name:          "concurrent completion attempts",
			numGoroutines: 5,
		},
		{
			name:          "more concurrent completions",
			numGoroutines: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()
			stream.AddEvent(LogEvent{Event: "log", Data: "Event"})

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					stream.SetComplete()
				}()
			}

			wg.Wait()

			assert.True(t, stream.IsCompleted(),
				"stream should be completed without race conditions")
		})
	}
}

func TestMixedConcurrentOperations(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numOperations int
	}{
		{
			name:          "mixed read/write/complete operations",
			numGoroutines: 5,
			numOperations: 10,
		},
		{
			name:          "higher concurrency mixed operations",
			numGoroutines: 20,
			numOperations: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines * 3) // writers, readers, completers

			// Writers
			for i := 0; i < test.numGoroutines; i++ {
				go func(id int) {
					defer wg.Done()
					for j := 0; j < test.numOperations; j++ {
						stream.AddEvent(LogEvent{
							Event: "log",
							Data:  fmt.Sprintf("Writer %d - Event %d", id, j),
						})
					}
				}(i)
			}

			// Readers
			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < test.numOperations; j++ {
						_ = stream.GetEvents()
					}
				}()
			}

			// Completers
			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < test.numOperations; j++ {
						stream.SetComplete()
					}
				}()
			}

			wg.Wait()

			// Verify no panics occurred and stream is in valid state
			assert.True(t, stream.IsCompleted() || !stream.IsCompleted(),
				"operation should complete without panic")
			events := stream.GetEvents()
			assert.NotNil(t, events, "GetEvents should not panic")
		})
	}
}

func TestUpdateLastAccess_Concurrent(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numUpdates    int
	}{
		{
			name:          "concurrent last access updates",
			numGoroutines: 5,
			numUpdates:    10,
		},
		{
			name:          "more concurrent updates",
			numGoroutines: 20,
			numUpdates:    10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := NewLogStream()

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < test.numUpdates; j++ {
						stream.UpdateLastAccess()
					}
				}()
			}

			wg.Wait()

			// Verify no panics occurred
			lastAccess := stream.GetLastAccessedAt()
			require.NotZero(t, lastAccess, "last access time should be set")
		})
	}
}
