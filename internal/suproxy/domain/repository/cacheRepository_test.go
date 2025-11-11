package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCacheRepository(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCacheRepository()

	tests := []struct {
		name       string
		setup      func()
		key        string
		wantHit    bool
		wantValue  []byte
		wait       time.Duration
		performDel bool
	}{
		{
			name: "miss on empty cache",
			setup: func() {
				// nothing
			},
			key:       "unknown",
			wantHit:   false,
			wantValue: nil,
		},
		{
			name: "hit after set",
			setup: func() {
				_ = cache.Set(ctx, "foo", []byte("bar"), time.Second)
			},
			key:       "foo",
			wantHit:   true,
			wantValue: []byte("bar"),
		},
		{
			name: "miss after expiry",
			setup: func() {
				_ = cache.Set(ctx, "temp", []byte("gone"), 100*time.Millisecond)
			},
			key:       "temp",
			wait:      200 * time.Millisecond,
			wantHit:   false,
			wantValue: nil,
		},
		{
			name: "miss after delete",
			setup: func() {
				_ = cache.Set(ctx, "del", []byte("x"), time.Second)
			},
			key:        "del",
			wantHit:    false,
			wantValue:  nil,
			performDel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset cache state between tests
			cache = NewMemoryCacheRepository()

			// Setup for this case
			tt.setup()

			// optional wait (for expiry tests)
			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}

			// optional delete
			if tt.performDel {
				_ = cache.Delete(ctx, tt.key)
			}

			got, hit, err := cache.Get(ctx, tt.key)
			assert.NoError(t, err, "Get should not return error")
			assert.Equal(t, tt.wantHit, hit, "hit mismatch")
			assert.Equal(t, tt.wantValue, got, "value mismatch")
		})
	}
}
