package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type TestcaseStorageService interface {
	SaveTestCase(context.Context, *entity.TestCase) error
}

type testcaseStorageService struct {
	logger *slog.Logger
	repo   repository.TestCaseStorageRepository
}

func NewTestcaseStorageService(logger *slog.Logger, repo repository.TestCaseStorageRepository) (TestcaseStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &testcaseStorageService{logger: logger, repo: repo}, nil
}

func (t *testcaseStorageService) SaveTestCase(ctx context.Context, testCase *entity.TestCase) error {
	return t.repo.Create(ctx, testCase)
}
