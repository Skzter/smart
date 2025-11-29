package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

// nolint: dupl
func TestNewTestcaseStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := &mocks.MockTestCaseStorageRepository{}
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
			svc, err := NewTestcaseStorageService(test.logger, test.repo)
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
			context:   nil,
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: nil,
			wantErr:   true,
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
			mockRepo := &mocks.MockTestCaseStorageRepository{}
			var key string
			if test.createErr == nil {
				key = "dummy-key"
			}
			mockRepo.EXPECT().Create(test.context, test.testCase, test.userId).Return(key, test.createErr)

			svc, err := NewTestcaseStorageService(logger, mockRepo)
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
