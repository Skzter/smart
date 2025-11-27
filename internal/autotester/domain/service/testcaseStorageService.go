package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseStorageService provides an interface to persist TestCase entities.
type TestcaseStorageService interface {
	// SaveTestCase persists the provided TestCase entity into the storage.
	// Returns an error if the operation fails.
	SaveTestcase(ctx context.Context, testcase *entity.TestCase, userId string) (string, error)
}

// testcaseStorageService implements the TestcaseStorageService interface
// and provides logic for storing TestCase entities via the underlying repository.
type testcaseStorageService struct {
	logger *slog.Logger
	repo   repository.TestcaseStorageRepository
}

// NewTestcaseStorageService creates a new TestcaseStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewTestcaseStorageService(logger *slog.Logger, repo repository.TestcaseStorageRepository) (TestcaseStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &testcaseStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveTestCase saves the given TestCase entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (t *testcaseStorageService) SaveTestcase(ctx context.Context, testcase *entity.TestCase, userId string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return "", err
	}

	key, err := t.repo.Create(ctx, testcase, userId)
	if err != nil {
		t.logger.Error("failed to save testcase",
			slog.String("testID", testcase.TestID),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	t.logger.Debug("testcase successfully saved",
		slog.String("testID", testcase.TestID),
	)
	return key, nil
}
