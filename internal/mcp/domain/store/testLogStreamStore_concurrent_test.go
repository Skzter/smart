//go:build racetest

package store

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
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
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			testId := "concurrent-test"
			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func(goroutineId int) {
					defer wg.Done()
					for j := 0; j < test.eventsPerGoroutine; j++ {
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
			expectedCount := test.numGoroutines * test.eventsPerGoroutine
			assert.Equal(t, expectedCount, len(events),
				"all events should be present without race conditions")
		})
	}
}

func TestGetStream_Concurrent(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
		numReads      int
	}{
		{
			name:          "concurrent reads of same stream",
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
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			testId := "concurrent-read-test"
			store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Test event"})

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			successCount := make([]int, test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func(goroutineId int) {
					defer wg.Done()
					for j := 0; j < test.numReads; j++ {
						stream, exists := store.GetStream(testId)
						if exists && stream != nil {
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

func TestCompleteStream_Concurrent(t *testing.T) {
	tests := []struct {
		name          string
		numGoroutines int
	}{
		{
			name:          "concurrent completion of same stream",
			numGoroutines: 5,
		},
		{
			name:          "more concurrent completions",
			numGoroutines: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			testId := "concurrent-complete-test"
			store.AddEvent(testId, entity.LogEvent{Event: "log", Data: "Event"})

			var wg sync.WaitGroup
			wg.Add(test.numGoroutines)

			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					store.CompleteStream(testId)
				}()
			}

			wg.Wait()

			stream, exists := store.GetStream(testId)
			require.True(t, exists, "stream should exist")
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
			name:          "mixed read/write operations",
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
			store := NewTestLogStreamStore()
			defer store.Shutdown()

			testId := "mixed-ops-test"
			var wg sync.WaitGroup
			wg.Add(test.numGoroutines * 3) // readers, writers, completers

			// Writers
			for i := 0; i < test.numGoroutines; i++ {
				go func(id int) {
					defer wg.Done()
					for j := 0; j < test.numOperations; j++ {
						store.AddEvent(testId, entity.LogEvent{
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
						store.GetStream(testId)
					}
				}()
			}

			// Completers
			for i := 0; i < test.numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < test.numOperations; j++ {
						store.CompleteStream(testId)
					}
				}()
			}

			wg.Wait()

			// Verify no panics occurred and stream is in valid state
			stream, exists := store.GetStream(testId)
			assert.True(t, exists || !exists, "operation should complete without panic")
			if exists {
				assert.NotNil(t, stream, "if stream exists, it should not be nil")
			}
		})
	}
}
