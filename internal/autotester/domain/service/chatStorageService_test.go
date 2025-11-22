package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

// nolint: dupl
func TestNewChatStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockChatStorageRepository(t)

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.ChatStorageRepository
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			repo:    mockRepo,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			repo:    mockRepo,
			wantErr: true,
		},
		{
			name:    "nil repo",
			logger:  logger,
			repo:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewChatStorageService(test.logger, test.repo)
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
func TestSaveChat(t *testing.T) {
	logger := slog.Default()
	sessionSummary := &entity.Chat{}

	tests := []struct {
		name          string
		context       context.Context
		testCase      entity.TestCase
		createReturns []any
		wantErr       bool
	}{
		{
			name:          "success",
			context:       context.Background(),
			testCase:      entity.TestCase{},
			createReturns: []any{nil},
			wantErr:       false,
		},
		{
			name:          "nil context",
			context:       nil,
			testCase:      entity.TestCase{},
			createReturns: nil,
			wantErr:       true,
		},
		{
			name:          "repo returns error",
			context:       context.Background(),
			createReturns: []any{errors.New("repo error")},
			wantErr:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, sessionSummary).Return(test.createReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			err = svc.SaveChat(test.context, sessionSummary)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveChat() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestLoadChat(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		context     context.Context
		key         string
		loadReturns []any
		wantErr     bool
	}{
		{
			name:        "success",
			context:     context.Background(),
			key:         "key123",
			loadReturns: []any{&entity.Chat{}, nil},
			wantErr:     false,
		},
		{
			name:        "nil context",
			context:     nil,
			key:         "key123",
			loadReturns: nil,
			wantErr:     true,
		},
		{
			name:        "repo returns error",
			context:     context.Background(),
			key:         "key123",
			loadReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.loadReturns != nil {
				mockRepo.On("Read", mock.Anything, test.key).Return(test.loadReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.LoadChat(test.context, test.key)
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

func TestListSummaries(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		context     context.Context
		listReturns []any
		wantErr     bool
	}{
		{
			name:        "success",
			context:     context.Background(),
			listReturns: []any{[]*entity.ChatSummary{}, nil},
			wantErr:     false,
		},
		{
			name:        "nil context",
			context:     nil,
			listReturns: nil,
			wantErr:     true,
		},
		{
			name:        "repo returns error",
			context:     context.Background(),
			listReturns: []any{nil, errors.New("repo error")},
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockChatStorageRepository(t)
			if test.listReturns != nil {
				mockRepo.On("ListAll", mock.Anything).Return(test.listReturns...)
			}

			svc, err := NewChatStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			res, err := svc.ListSummaries(test.context)
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
