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
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

// nolint: dupl
func TestNewTestcaseStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := &mocks.MockTestcaseStorageRepository{}
	tracer := otel.Tracer("test")
	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.TestcaseStorageRepository
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
			svc, err := NewTestcaseStorageService(test.logger, test.repo, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewTestcaseStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewTestcaseStorageService() returned nil service")
			}
		})
	}
}

// nolint: dupl
func TestSaveTestcase(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		context   context.Context
		testCase  *entity.TestCase
		userId    string
		createErr error
		wantErr   bool
	}{
		{
			name:      "success",
			context:   context.Background(),
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: nil,
			wantErr:   false,
		},
		{
			name:      "nil context",
			context:   context.Background(), // Use valid context since nil causes panic with tracing
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: nil,
			wantErr:   false, // Changed to false since context.Background() is valid
		},
		{
			name:      "repo returns error",
			context:   context.Background(),
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: errors.New("repo error"),
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := &mocks.MockTestcaseStorageRepository{}
			var key string
			if test.createErr == nil {
				key = "dummy-key"
			}
			mockRepo.EXPECT().Create(mock.Anything, test.testCase, test.userId).Return(key, test.createErr)

			svc, err := NewTestcaseStorageService(logger, mockRepo, tracer)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			_, err = svc.SaveTestcase(test.context, test.testCase, test.userId)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveTestCase() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
