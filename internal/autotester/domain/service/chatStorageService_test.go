package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	servmocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockChatStorageRepository(t)
	mockValidator := servmocks.NewMockValidator(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil assertion error",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewChatStorageService(test.logger, mockRepo, mockValidator, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewSessionSummaryStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewSessionSummaryStorageService() returned nil service")
			}
		})
	}
}

// nolint: dupl
func TestChatStorageSaveChat(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name          string
		createReturns []any
		validRetuns   []any
		wantErr       bool
		chat          *entity.Chat
	}{
		{
			name:          "success",
			createReturns: []any{nil},
			validRetuns:   []any{nil},
			wantErr:       false,
			chat:          &entity.Chat{},
		},
		{
			name:    "nil assert error",
			wantErr: true,
			chat:    nil,
		},
		{
			name:        "validation error",
			wantErr:     true,
			validRetuns: []any{errors.New("err")},
			chat:        &entity.Chat{},
		},
		{
			name:          "repo returns error",
			createReturns: []any{errors.New("repo error")},
			validRetuns:   []any{nil},
			wantErr:       true,
			chat:          &entity.Chat{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.chat).Return(test.createReturns...)
			}
			if test.validRetuns != nil {
				mockVal.On("ValidateChat", mock.Anything, test.chat).Return(test.validRetuns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			err = svc.SaveChat(context.Background(), test.chat)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveChat() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestChatStorageLoadChat(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name         string
		chatid       string
		loadReturns  []any
		validReturns []any
		ctx          context.Context
		wantErr      bool
	}{
		{
			name:         "success",
			chatid:       "chat123",
			loadReturns:  []any{&entity.Chat{}, nil},
			validReturns: []any{nil},
			ctx:          context.Background(),
			wantErr:      false,
		},
		{
			name:    "nil assert failed",
			ctx:     nil,
			wantErr: true,
		},
		{
			name:    "invalid chatId",
			chatid:  "",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:        "repo returns error",
			chatid:      "chat123",
			loadReturns: []any{nil, errors.New("repo error")},
			ctx:         context.Background(),
			wantErr:     true,
		},
		{
			name:         "repo returns invalid chat",
			chatid:       "chat123",
			loadReturns:  []any{&entity.Chat{}, nil},
			validReturns: []any{errors.New("err")},
			ctx:          context.Background(),
			wantErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.loadReturns != nil {
				mockRepo.On("Read", mock.Anything, test.chatid).Return(test.loadReturns...)
			}
			if test.validReturns != nil {
				mockVal.On("ValidateChat", mock.Anything, &entity.Chat{}).Return(test.validReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo, mockVal, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadChat(test.ctx, test.chatid)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestLoadSummaries(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	orderedResult := []*entity.ChatSummary{{UpdatedAt: time.Unix(200, 0), Groups: []string{"1", "2"}}, {UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}}}

	tests := []struct {
		name        string
		groups      []string
		listReturns []any
		expected    []*entity.ChatSummary
		wantErr     bool
		ctx         context.Context
	}{
		{
			name:        "success, no groups",
			listReturns: []any{orderedResult, nil},
			expected:    orderedResult,
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{},
		},
		{
			name:        "success, filtered by Group groups",
			listReturns: []any{append(orderedResult, &entity.ChatSummary{UpdatedAt: time.Unix(1, 0), Groups: []string{"0"}}), nil},
			expected:    orderedResult,
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{"1"},
		},
		{
			name:    "nil assert error",
			wantErr: true,
			ctx:     nil,
		},
		{
			name:    "group assert error",
			wantErr: true,
			ctx:     context.Background(),
			groups:  []string{""},
		},
		{
			name:        "inverse order",
			listReturns: []any{[]*entity.ChatSummary{{UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}}, {UpdatedAt: time.Unix(200, 0), Groups: []string{"1", "2"}}}, nil},
			expected:    orderedResult,
			wantErr:     false,
			ctx:         context.Background(),
			groups:      []string{},
		},
		{
			name: "secondary sort by ChatId when UpdatedAt equal",
			listReturns: []any{[]*entity.ChatSummary{
				{ChatId: "chat-z", UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}},
				{ChatId: "chat-a", UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}},
				{ChatId: "chat-m", UpdatedAt: time.Unix(200, 0), Groups: []string{"1"}},
			}, nil},
			expected: []*entity.ChatSummary{
				{ChatId: "chat-m", UpdatedAt: time.Unix(200, 0), Groups: []string{"1"}},
				{ChatId: "chat-a", UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}},
				{ChatId: "chat-z", UpdatedAt: time.Unix(100, 0), Groups: []string{"1"}},
			},
			wantErr: false,
			ctx:     context.Background(),
			groups:  []string{},
		},
		{
			name:        "repo returns error",
			listReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
			ctx:         context.Background(),
			groups:      []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.listReturns != nil {
				mockRepo.On("ListAll", mock.Anything).Return(test.listReturns...)
			}

			val := servmocks.NewMockValidator(t)

			svc, err := NewChatStorageService(logger, mockRepo, val, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadSummaries(test.ctx, test.groups...)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.Nil(t, err)
				if test.expected != nil {
					assert.Equal(t, test.expected, res)
				} else {
					assert.NotNil(t, res)
				}
			}
		})
	}
}
