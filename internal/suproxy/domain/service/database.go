package service

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// DatabaseService defines an interface for database operations.
type DatabaseService interface {
	// SaveDbEntry saves a database entry.
	SaveDbEntry(context.Context, entity.DatabaseEntry) error
	// GetAllKeys retrieves all keys from the database.
	GetAllKeys(ctx context.Context) ([]string, error)
}

// databaseService provides access to the database through the configured repository.
type databaseService struct {
	logger *slog.Logger
	repo   repository.DatabaseRepository
	tracer trace.Tracer
}

// NewDatabaseService creates a new instance of DatabaseService.
func NewDatabaseService(logger *slog.Logger, repo repository.DatabaseRepository,
	tracer trace.Tracer,
) (DatabaseService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, fmt.Errorf("repo cannot be nil, %w", err)
	}
	return &databaseService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

// SaveDbEntry saves a database entry by calling the repository's CreateRequest method.
func (d *databaseService) SaveDbEntry(ctx context.Context, request entity.DatabaseEntry) error {
	ctx, span := d.tracer.Start(ctx, "databaseService.SaveDbEntry")
	defer span.End()

	if err := assert.NotNil(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "context validation failed")
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := d.repo.CreateRequest(ctx, request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save request")
		return fmt.Errorf("failed to save request: %w", err)
	}
	d.logger.Debug("Request saved successfully", "request", request)
	span.SetStatus(codes.Ok, "")
	return nil
}

// GetAllKeys retrieves all keys from the database.
func (d *databaseService) GetAllKeys(ctx context.Context) ([]string, error) {
	ctx, span := d.tracer.Start(ctx, "databaseService.GetAllKeys")
	defer span.End()

	if err := assert.NotNil(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "context validation failed")
		return nil, fmt.Errorf("context cannot be nil")
	}
	keys, err := d.repo.ListAllKeys(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get all keys")
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}
	span.SetStatus(codes.Ok, "")
	return keys, nil
}
