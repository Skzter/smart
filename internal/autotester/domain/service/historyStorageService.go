package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type HistoryStorageService interface {
	SaveHistory(ctx context.Context, summary entity.SessionSummary) error
}
type historyStorageService struct {
	logger *slog.Logger
	repo   repository.SessionSummaryStorageRepository
}

func NewHistoryStorageService(logger *slog.Logger, repo repository.SessionSummaryStorageRepository) (HistoryStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &historyStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

// Method called by the handler
func (s *historyStorageService) SaveHistory(ctx context.Context, summary entity.SessionSummary) error {
	if err := assert.NotNil(ctx, summary); err != nil {
		return err
	}
	return s.repo.Create(ctx, &summary)
}
