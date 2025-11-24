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
	tracer trace.Tracer
}

// NewHistoryStorageService creates a new HistoryStorageService instance.
// Returns the service or an error if any of the arguments are nil.
func NewHistoryStorageService(
	logger *slog.Logger,
	repo repository.SessionSummaryStorageRepository,
	tracer trace.Tracer,
) (HistoryStorageService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, err
	}

	return &historyStorageService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

// SaveHistory saves the given SessionSummary entity using the configured repository.
// Validates the input context and returns an error if it is nil or if the repository operation fails.
func (s *historyStorageService) SaveHistory(ctx context.Context, summary *entity.SessionSummary) error {
	ctx, span := s.tracer.Start(ctx, "historyStorageService.SaveHistory")
	defer span.End()

	if err := assert.NotNil(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "context validation failed")
		return err
	}
	err := s.repo.Create(ctx, summary)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save history")
		return err
	}
	span.SetStatus(codes.Ok, "")
	return nil
}
