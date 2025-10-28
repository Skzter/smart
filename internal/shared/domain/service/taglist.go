package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistService defines the interface for managing taglists.
type TaglistService interface {
	// StoreTaglist updates stored taglist
	StoreTaglist(context.Context, []string) error
	// GetTaglist retrieves Taglist from storage
	GetTaglist(ctx context.Context) ([]string, error)
}

// taglistService provides access to the database through the configured repository.
type taglistService struct {
	logger *slog.Logger
	repo   repository.TaglistRepository
}

// NewTaglistService creates a new instance of TaglistService.
func NewTaglistService(logger *slog.Logger, repo repository.TaglistRepository) (TaglistService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, fmt.Errorf("repo cannot be nil, %w", err)
	}
	return &taglistService{
		logger: logger,
		repo:   repo,
	}, nil
}

// StoreTaglist updates stored taglist
func (d *taglistService) StoreTaglist(ctx context.Context, tags []string) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	// Check if taglist exists
	exists, err := d.repo.TaglistExists(ctx)
	if err != nil {
		return fmt.Errorf("S3 Error: %w", err)
	}

	taglist := entity.TagListEntity{Tags: tags}
	// Create or update taglist
	if !exists {
		if err := d.repo.CreateTaglist(ctx, taglist); err != nil {
			return fmt.Errorf("failed to save taglist: %w", err)
		}
	} else {
		if err := d.repo.UpdateTaglist(ctx, taglist); err != nil {
			return fmt.Errorf("failed to save taglist: %w", err)
		}
	}

	d.logger.Debug("Taglist saved successfully", "request", taglist)
	return nil
}

// GetTaglist retrieves Taglist from storage
func (d *taglistService) GetTaglist(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil, %w", err)
	}

	list, err := d.repo.ReadTaglist(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get taglist, %w", err)
	}

	return list.Tags, nil
}
