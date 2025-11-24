package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	repository_intf "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

// nolint: dupl
func TestNewHistoryStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := &mocks.MockSessionSummaryStorageRepository{}
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository_intf.SessionSummaryStorageRepository
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
			svc, err := NewHistoryStorageService(test.logger, test.repo, tracer)
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
func TestSaveHistory(t *testing.T) {
	logger := slog.Default()
	sessionSummary := &entity.SessionSummary{}
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		context   context.Context
		testCase  entity.TestCase
		createErr error
		wantErr   bool
	}{
		{
			name:      "success",
			context:   context.Background(),
			testCase:  entity.TestCase{},
			createErr: nil,
			wantErr:   false,
		},
		{
			name:      "nil context",
			context:   nil,
			testCase:  entity.TestCase{},
			createErr: nil,
			wantErr:   true,
		},
		{
			name:      "repo returns error",
			context:   context.Background(),
			createErr: errors.New("repo error"),
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := &mocks.MockSessionSummaryStorageRepository{}
			mockRepo.On("Create", mock.Anything, sessionSummary).Return(test.createErr)

			svc, err := NewHistoryStorageService(logger, mockRepo, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			err = svc.SaveHistory(test.context, sessionSummary)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveHistory() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
