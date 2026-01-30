package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistStorage defines the interface for managing taglists.
type TaglistStorage interface {
	// StoreTaglist updates stored taglist
	StoreTaglist(context.Context, *entity.TagList) error
	// GetTaglist retrieves Taglist from storage
	GetTaglist(ctx context.Context) (*entity.TagList, error)
}

// taglistStorage provides access to the database through the configured repository.
type taglistStorage struct {
	logger *slog.Logger
	repo   repository.TaglistStorage
}

// NewTaglistStorage creates a new instance of TaglistService.
func NewTaglistStorage(logger *slog.Logger, repo repository.TaglistStorage) (TaglistStorage, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}
	return &taglistStorage{
		logger: logger,
		repo:   repo,
	}, nil
}

// StoreTaglist updates stored taglist
func (d *taglistStorage) StoreTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx); err != nil {
		d.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return sharedErrors.ErrGeneration
	}
	// Check if taglist exists
	exists, err := d.repo.TaglistExists(ctx)
	if err != nil {
		d.logger.Error(fmt.Sprintf("S3 Error: %s", err))
		return sharedErrors.ErrInternalServer
	}

	// Create or update taglist
	if !exists {
		err = d.repo.CreateTaglist(ctx, taglist)
	} else {
		err = d.repo.UpdateTaglist(ctx, taglist)
	}
	if err != nil {
		d.logger.Error(fmt.Sprintf("failed to save taglist: %s", err))
		return sharedErrors.ErrInternalServer
	}

	d.logger.Debug("Taglist saved successfully", "taglist", taglist)
	return nil
}

// GetTaglist retrieves Taglist from storage
func (d *taglistStorage) GetTaglist(ctx context.Context) (*entity.TagList, error) {
	if err := assert.NotNil(ctx); err != nil {
		d.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return nil, sharedErrors.ErrInternalServer
	}

	list, err := d.repo.ReadTaglist(ctx)

	if err != nil {
		return nil, err
	}

	return list, nil
}
