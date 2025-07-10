package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

type HistoryStorageService interface {
	SaveHistory(ctx context.Context, summary entity.SessionSummary) error
}
type historyStorageService struct {
	logger *slog.Logger
	repo   repository.SessionSummaryStorageRepository
}

func NewHistoryStorageService(logger *slog.Logger, repo repository.SessionSummaryStorageRepository) HistoryStorageService {
	if logger == nil {
		logger = slog.Default()
	}
	if repo == nil {
		log.Fatal("SessionSummaryStorageRepository must not be nil")
	}
	return &historyStorageService{
		logger: logger,
		repo:   repo,
	}
}

// Method called by the handler
func (s *historyStorageService) SaveHistory(ctx context.Context, summary entity.SessionSummary) error {
	if context.Context. == nil {
		return fmt.Errorf("SessionSummaryStorageRepository is not set")
	}
	if err := s.repo.Create(ctx, &summary); err != nil {
		return err
	}
	return nil
}
