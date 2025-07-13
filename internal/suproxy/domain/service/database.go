package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

type DatabaseService interface {
	// SaveDbEntry saves a database entry.
	SaveDbEntry(context.Context, entity.DatabaseEntry) error
	// GetAllKeys retrieves all keys from the database.
	GetAllKeys(ctx context.Context) ([]string, error)
}

type databaseService struct {
	logger *slog.Logger
	repo   repository.DatabaseRepository
}

// NewDatabaseService creates a new instance of DatabaseService.
func NewDatabaseService(logger *slog.Logger, repo repository.DatabaseRepository) (DatabaseService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, fmt.Errorf("repo cannot be nil, %w", err)
	}
	return &databaseService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveDbEntry saves a database entry by calling the repository's CreateRequest method.
func (d *databaseService) SaveDbEntry(ctx context.Context, request entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := d.repo.CreateRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}
	d.logger.Debug("Request saved successfully", "request", request)
	return nil
}

// GetAllKeys retrieves all keys from the database.
func (d *databaseService) GetAllKeys(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	keys, err := d.repo.ListAllKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}
	return keys, nil
}
