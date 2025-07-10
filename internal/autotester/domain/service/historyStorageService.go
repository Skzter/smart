package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HistoryStorageService provides an interface to persist SessionSummary entities.
type HistoryStorageService interface {
	// SaveHistory persists the provided SessionSummary entity into the storage.
	// Returns an error if the operation fails.
	SaveHistory(ctx context.Context, summary *entity.SessionSummary) error
}

// historyStorageService implements the HistoryStorageService interface
// and provides logic for storing SessionSummary entities via the underlying repository.
type historyStorageService struct {
	logger *slog.Logger
	repo   repository.SessionSummaryStorageRepository
}

// NewHistoryStorageService creates a new HistoryStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewHistoryStorageService(logger *slog.Logger, repo repository.SessionSummaryStorageRepository) (HistoryStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &historyStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveHistory saves the given SessionSummary entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (s *historyStorageService) SaveHistory(ctx context.Context, summary *entity.SessionSummary) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	return s.repo.Create(ctx, summary)
}
