package service

import (
	"context"
	"fmt"
	"log/slog"

	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
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
}

// NewDatabaseService creates a new instance of DatabaseService.
func NewDatabaseService(logger *slog.Logger, repo repository.DatabaseRepository) (DatabaseService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		logger.Error(fmt.Sprintf("repo cannot be nil, %s", err))
		return nil, sharedErrors.ErrInternalServer
	}
	return &databaseService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveDbEntry saves a database entry by calling the repository's CreateRequest method.
func (d *databaseService) SaveDbEntry(ctx context.Context, request entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		d.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return sharedErrors.ErrInternalServer
	}
	if err := d.repo.CreateRequest(ctx, request); err != nil {
		d.logger.Error(fmt.Sprintf("failed to save request: %s", err))
		return sharedErrors.ErrInternalServer
	}
	d.logger.Debug("Request saved successfully", "request", request)
	return nil
}

// GetAllKeys retrieves all keys from the database.
func (d *databaseService) GetAllKeys(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		d.logger.Error("context cannot be nil")
		return nil, fmt.Errorf("context cannot be nil")
	}
	keys, err := d.repo.ListAllKeys(ctx)
	if err != nil {
		d.logger.Error(fmt.Sprintf("failed to get all keys: %s", err))
		return nil, sharedErrors.ErrInternalServer
	}
	return keys, nil
}
