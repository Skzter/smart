package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

type DatabaseService interface {
	// SaveDbEntry saves a database entry.
	SaveDbEntry(context.Context, entity.DatabaseEntry) error
	// GetAllKeys retrieves all keys from the database.
	GetAllKeys(ctx context.Context) ([]string, error)
	// GetKey retrieves a key based on the provided tags.
	GetKeysForTags(ctx context.Context, tags []string) ([]string, error)
}

type databaseService struct {
	logger *slog.Logger
	repo   repository.DatabaseRepository
}

// NewDatabaseService creates a new instance of DatabaseService.
func NewDatabaseService(logger *slog.Logger, repo repository.DatabaseRepository) (DatabaseService, error) {
	if logger == nil || repo == nil {
		return nil, fmt.Errorf("logger and repository cannot be nil")
	}
	return &databaseService{
		logger: logger,
		repo:   repo,
	}, nil
}

// SaveDbEntry saves a database entry by calling the repository's CreateRequest method.
func (d *databaseService) SaveDbEntry(ctx context.Context, request entity.DatabaseEntry) error {
	if err := d.repo.CreateRequest(ctx, request); err != nil {
		return fmt.Errorf("failed to save request: %w", err)
	}
	d.logger.Debug("Request saved successfully", "request", request)
	return nil
}

// GetAllKeys retrieves all keys from the database.
func (d *databaseService) GetAllKeys(ctx context.Context) ([]string, error) {
	keys, err := d.repo.ListKeysFromFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all keys: %w", err)
	}
	return keys, nil
}

// GetKeysForTags retrieves keys based on the provided tags.
func (d *databaseService) GetKeysForTags(ctx context.Context, tags []string) ([]string, error) {
	keys, err := d.GetAllKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	expectedTags := strings.Join(tags, "-") + "-"
	var matchedKeys []string
	for _, key := range keys {
		if strings.HasPrefix(key, expectedTags) {
			matchedKeys = append(matchedKeys, key)
		}
	}
	if len(matchedKeys) == 0 {
		return nil, fmt.Errorf("no key found for tags: %v", tags)
	}
	return matchedKeys, nil
}
