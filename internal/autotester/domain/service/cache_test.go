package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

func TestNewCacheService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tracer := otel.Tracer("test")
	tests := []struct {
		name      string
		logger    *slog.Logger
		config    *config.Autotester
		cacheRepo repository.Cache
		wantError bool
	}{
		{
			name:      "success - service gets build",
			logger:    logger,
			config:    cfg,
			cacheRepo: sharedMocks.NewMockCache(t),
			wantError: false,
		},
		{
			name:      "error - nil params",
			logger:    nil,
			config:    nil,
			cacheRepo: nil,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewCacheService(tc.config, tc.logger, tc.cacheRepo, tracer)
			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, srv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, srv)
			}
		})
	}
}

func TestLookUp(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tracer := otel.Tracer("test")
	chatId := "user456"
	chat := entity.NewChat(chatId, nil)
	chatByte, _ := json.Marshal(chat)

	tests := []struct {
		name              string
		ctx               context.Context
		chatId            string
		mockCacheResponse []any
		cacheHit          bool
		wantChat          *entity.Chat
		wantErr           bool
	}{
		{
			name:    "error - nil ctx",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "error - empty chat id",
			ctx:     t.Context(),
			chatId:  "",
			wantErr: true,
		},
		{
			name:              "error - cache repo returned error",
			ctx:               t.Context(),
			chatId:            "123",
			mockCacheResponse: []any{nil, false, fmt.Errorf("cache failed")},
			wantErr:           true,
		},
		{
			name:              "error - json unmarshal failed",
			ctx:               t.Context(),
			chatId:            "123",
			mockCacheResponse: []any{nil, true, nil},
			wantErr:           true,
		},
		{
			name:              "success - cache miss, no error, no chat",
			ctx:               t.Context(),
			chatId:            "123",
			mockCacheResponse: []any{nil, false, nil},
			wantChat:          nil,
			wantErr:           false,
		},
		{
			name:              "success - cache hit, no error, chat",
			ctx:               t.Context(),
			chatId:            chatId,
			mockCacheResponse: []any{chatByte, true, nil},
			cacheHit:          true,
			wantChat:          chat,
			wantErr:           false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCache := sharedMocks.NewMockCache(t)
			cacheSrv := &cache{
				logger:    logger,
				config:    cfg,
				cacheRepo: mockCache,
				tracer:    tracer,
			}

			if tc.mockCacheResponse != nil {
				mockCache.On("Get", mock.Anything, tc.chatId).Return(tc.mockCacheResponse...)
			}
			chat, err := cacheSrv.LookUp(tc.ctx, tc.chatId)
			t.Logf("tc setup\nhit: %t, wantErr: %t\n", tc.cacheHit, tc.wantErr)
			t.Logf("Chat is %v and error %v and wantChat %v\n", chat, err, tc.wantChat)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, chat)
				assert.Equal(t, tc.wantChat, chat)
			} else {
				if !tc.cacheHit {
					assert.NoError(t, err)
					assert.Nil(t, chat)
					assert.Equal(t, tc.wantChat, chat)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, chat)
					assert.Equal(t, tc.wantChat, chat)
				}
			}
		})
	}
}
