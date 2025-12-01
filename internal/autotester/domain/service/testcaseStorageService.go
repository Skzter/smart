package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseStorageService provides an interface to persist TestCase entities.
type TestcaseStorageService interface {
	// SaveTestCase persists the provided TestCase entity into the storage.
	// Returns an error if the operation fails.
	SaveTestCase(context.Context, *entity.TestCase) error
}

// testcaseStorageService implements the TestcaseStorageService interface
// and provides logic for storing TestCase entities via the underlying repository.
type testcaseStorageService struct {
	logger *slog.Logger
	repo   repository.TestCaseStorageRepository
	tracer trace.Tracer
}

// NewTestcaseStorageService creates a new TestcaseStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewTestcaseStorageService(
	logger *slog.Logger,
	repo repository.TestCaseStorageRepository,
	tracer trace.Tracer,
) (TestcaseStorageService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, err
	}

	return &testcaseStorageService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

// SaveTestCase saves the given TestCase entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (t *testcaseStorageService) SaveTestCase(ctx context.Context, testCase *entity.TestCase) error {
	ctx, span := t.tracer.Start(ctx, "testcaseStorageService.SaveTestCase")
	defer span.End()

	if err := assert.NotNil(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "context validation failed")
		return err
	}
	err := t.repo.Create(ctx, testCase)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save test case")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
