package service

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistSync defines interface for syncing taglist.
type TaglistSync interface {
	// SyncTaglist syncs stored taglist.
	SyncTaglist(context.Context, *sharedEntity.TagList) error
	// GetTaglist gets Taglist
	GetCurrentTaglist() *sharedEntity.TagList
}

type taglistSync struct {
	logger         *slog.Logger
	taglistService service.TaglistStorage
	tagList        *sharedEntity.TagList
	mutex          sync.Mutex
}

// NewTaglistSync syncs taglist.
func NewTaglistSync(logger *slog.Logger, taglistService service.TaglistStorage) (TaglistSync, error) {
	if err := assert.NotNil(logger, taglistService); err != nil {
		return nil, err
	}

	taglist, err := taglistService.GetTaglist(context.Background())
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	// mutex requires no initialization
	return &taglistSync{
		logger:         logger,
		taglistService: taglistService,
		tagList:        taglist,
	}, nil
}

// SyncTaglist syncs given taglist with in-memory and pushes new taglist to s3
func (tls *taglistSync) SyncTaglist(ctx context.Context, taglist *sharedEntity.TagList) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	if len(taglist.Tags) == 0 {
		return errors.New("empty taglist")
	}

	tls.mutex.Lock()
	defer tls.mutex.Unlock()
	lenghtList := len(tls.tagList.Tags)
	for _, tag := range taglist.Tags {
		tls.logger.Error("geht in schleife rein")
		if !slices.Contains(tls.tagList.Tags, tag) {
			tls.tagList.Tags = append(tls.tagList.Tags, tag)
		}
	}

	if lenghtList != len(tls.tagList.Tags) {
		tls.logger.Info("UPDATES")
		err := tls.taglistService.StoreTaglist(ctx, tls.tagList)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetTaglist gets Taglist
func (tls *taglistSync) GetCurrentTaglist() *sharedEntity.TagList {
	newTags := append([]sharedEntity.Tag(nil), tls.tagList.Tags...)
	tagList := sharedEntity.TagList{Tags: newTags}
	return &tagList
}
