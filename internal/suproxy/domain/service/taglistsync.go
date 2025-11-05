package service

import (
	"context"
	"log/slog"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistSyncService defines interface for syncing taglist.
type TaglistSyncService interface {
	// SyncTaglist syncs stored taglist.
	SyncTaglist(context.Context, []string) error
}

type taglistSync struct {
	logger         *slog.Logger
	taglistService service.TaglistStorage
	tagList        []string
}

// NewTaglistSync syncs taglist.
func NewTaglistSync(logger *slog.Logger, taglistService service.TaglistStorage) (TaglistSyncService, error) {
	if err := assert.NotNil(logger, taglistService); err != nil {
		return nil, err
	}
	taglist, err := taglistService.GetTaglist(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return &taglistSync{
		logger:         logger,
		taglistService: taglistService,
		tagList:        taglist,
	}, nil
}

// SyncTaglist syncs stored taglist.
func (tls *taglistSync) SyncTaglist(ctx context.Context, tags []string) error {
	return nil
}
