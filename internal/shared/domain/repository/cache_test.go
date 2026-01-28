package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/config"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks/repository"
)

const testRedisKey = "test-key"

func newCacheTestLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func setupCacheMocks(t *testing.T) (*redisCache, *mocks.MockRedisClient) {
	logger := newCacheTestLogger()
	mockClient := mocks.NewMockRedisClient(t)

	cache := &redisCache{
		log:    logger,
		rdb:    mockClient,
		tracer: otel.Tracer("test"),
	}

	return cache, mockClient
}

func TestNewRedisCache(t *testing.T) {
	tracer := otel.Tracer("test")
	tests := []struct {
		name     string
		logger   *slog.Logger
		cfg      *config.RedisConfig
		wantErr  bool
		wantAddr string
	}{
		{
			name:    "nil cfg → error",
			logger:  newCacheTestLogger(),
			cfg:     nil,
			wantErr: true,
		},
		{
			name:    "nil logger → error",
			logger:  nil,
			cfg:     &config.RedisConfig{},
			wantErr: true,
		},
		{
			name:   "full valid redis config",
			logger: newCacheTestLogger(),
			cfg: &config.RedisConfig{
				Addr:     "localhost:9999",
				Password: "pw",
				Db:       2,
				Protocol: 3,
			},
			wantErr:  false,
			wantAddr: "localhost:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := NewRedisCache(tt.logger, tt.cfg, tracer)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cache)

			if tt.wantAddr != "" {
				rc := cache.(*redisCache)
				realClient, ok := rc.rdb.(*redis.Client)
				assert.True(t, ok)
				opts := realClient.Options()
				assert.Equal(t, tt.wantAddr, opts.Addr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		key       string
		mockValue string
		mockErr   error
		wantHit   bool
		wantErr   bool
		wantValue []byte
	}{
		{
			name:      "ctx nil → error",
			ctx:       nil,
			key:       testRedisKey,
			wantErr:   true,
			wantHit:   false,
			wantValue: nil,
		},
		{
			name:      "hit",
			ctx:       context.Background(),
			key:       testRedisKey,
			mockValue: "123",
			mockErr:   nil,
			wantHit:   true,
			wantErr:   false,
			wantValue: []byte("123"),
		},
		{
			name:      "miss",
			ctx:       context.Background(),
			key:       testRedisKey,
			mockValue: "",
			mockErr:   redis.Nil,
			wantHit:   false,
			wantErr:   false,
			wantValue: nil,
		},
		{
			name:      "redis error",
			ctx:       context.Background(),
			key:       testRedisKey,
			mockValue: "",
			mockErr:   errors.New("boom"),
			wantHit:   false,
			wantErr:   true,
			wantValue: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, mockClient := setupCacheMocks(t)

			if tt.ctx != nil {
				cmd := redis.NewStringCmd(tt.ctx)
				cmd.SetVal(tt.mockValue)
				cmd.SetErr(tt.mockErr)

				mockClient.On("Get", mock.Anything, tt.key).
					Return(cmd).
					Once()
			}

			val, hit, err := cache.Get(tt.ctx, tt.key)

			assert.Equal(t, tt.wantHit, hit)
			assert.Equal(t, tt.wantValue, val)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestSet(t *testing.T) {
	value := []byte("hello")
	tests := []struct {
		name    string
		ctx     context.Context
		key     string
		value   []byte
		ttl     time.Duration
		mockErr error
		wantErr bool
	}{
		{
			name:    "ctx nil → error",
			ctx:     nil,
			key:     testRedisKey,
			value:   value,
			ttl:     time.Second,
			wantErr: true,
		},
		{
			name:    "success",
			ctx:     context.Background(),
			key:     testRedisKey,
			value:   value,
			ttl:     time.Second,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "redis error",
			ctx:     context.Background(),
			key:     testRedisKey,
			value:   value,
			ttl:     time.Second,
			mockErr: errors.New("fail"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, mockClient := setupCacheMocks(t)

			if tt.ctx != nil {
				cmd := redis.NewStatusCmd(tt.ctx)
				cmd.SetErr(tt.mockErr)

				mockClient.
					On("Set", mock.Anything, tt.key, tt.value, tt.ttl).
					Return(cmd).
					Once()
			}

			err := cache.Set(tt.ctx, tt.key, tt.value, tt.ttl)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		key     string
		mockErr error
		wantErr bool
	}{
		{
			name:    "ctx nil → error",
			ctx:     nil,
			key:     testRedisKey,
			wantErr: true,
		},
		{
			name:    "success",
			ctx:     context.Background(),
			key:     testRedisKey,
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "redis error",
			ctx:     context.Background(),
			key:     testRedisKey,
			mockErr: errors.New("delete failed"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, mockClient := setupCacheMocks(t)

			if tt.ctx != nil {
				cmd := redis.NewIntCmd(tt.ctx)
				cmd.SetErr(tt.mockErr)

				mockClient.
					On("Del", mock.Anything, []string{tt.key}).
					Return(cmd).
					Once()
			}

			err := cache.Delete(tt.ctx, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestClose(t *testing.T) {
	cache, mockClient := setupCacheMocks(t)

	mockClient.On("Close").Return(nil).Once()

	err := cache.Close()

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
