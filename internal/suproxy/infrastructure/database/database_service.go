package database

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	domainRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
	domainService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// databaseService is the infrastructure implementation of the domain DatabaseService.
type databaseService struct {
	logger *slog.Logger
	repo   domainRepo.DatabaseRepository
	tracer trace.Tracer
}

// NewDatabaseService constructs the infrastructure implementation.
func NewDatabaseService(
	logger *slog.Logger,
	repo domainRepo.DatabaseRepository,
	tracer trace.Tracer,
) (domainService.DatabaseService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, fmt.Errorf("dependency cannot be nil: %w", err)
	}

	return &databaseService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

func (d *databaseService) SaveDbEntry(ctx context.Context, entry entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil: %w", err)
	}

	ctx, span := d.tracer.Start(ctx, "databaseService.SaveDbEntry")
	defer span.End()

	if err := d.repo.CreateRequest(ctx, entry); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save entry")
		return fmt.Errorf("failed to save entry: %w", err)
	}

	d.logger.Debug("database entry saved", "entry", entry)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (d *databaseService) GetAllKeys(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	ctx, span := d.tracer.Start(ctx, "databaseService.GetAllKeys")
	defer span.End()

	keys, err := d.repo.ListAllKeys(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list keys")
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return keys, nil
}
