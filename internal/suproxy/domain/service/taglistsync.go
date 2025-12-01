package service

import (
	"context"
	"fmt"
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
	}

	logger.Debug(fmt.Sprintf("Initial Taglist => %v", taglist))
	if taglist == nil || len(taglist.Tags) == 0 {
		taglist = sharedEntity.DefaultTagList()
		err := taglistService.StoreTaglist(context.Background(), taglist)
		if err != nil {
			logger.Error(err.Error())
		}
		logger.Debug(fmt.Sprintf("Updated Taglist => %v", taglist))
	}

	// mutex requires no initialization
	return &taglistSync{
		logger:         logger,
		taglistService: taglistService,
		tagList:        taglist,
	}, nil
}

// SyncTaglist synchronizes the current taglist with the provided taglist, ensuring all tags are up-to-date and stored.
func (tls *taglistSync) SyncTaglist(ctx context.Context, taglist *sharedEntity.TagList) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}

	tls.mutex.Lock()
	defer tls.mutex.Unlock()

	incomingTags := taglist.Tags

	allExist := true
	for _, t := range incomingTags {
		if !slices.Contains(tls.tagList.Tags, t) {
			allExist = false
			break
		}
	}
	if allExist {
		return nil
	}

	onlineTaglist, err := tls.taglistService.GetTaglist(ctx)
	if err != nil {
		tls.logger.Error("Failed to fetch taglist from S3", "error", err)
		return err
	}
	if len(onlineTaglist.Tags) == 0 {
		tls.logger.Warn("onlineTaglist is empty, proceeding with local taglist")
	}

	needsUpdate := false
	for _, tag := range onlineTaglist.Tags {
		if !slices.Contains(tls.tagList.Tags, tag) {
			tls.tagList.Tags = append(tls.tagList.Tags, tag)
			needsUpdate = true
		}
	}

	for _, t := range incomingTags {
		if !slices.Contains(tls.tagList.Tags, t) {
			tls.tagList.Tags = append(tls.tagList.Tags, t)
			needsUpdate = true
		}
	}

	if !needsUpdate {
		return nil
	}

	err = tls.taglistService.StoreTaglist(ctx, tls.tagList)
	if err != nil {
		tls.logger.Error("Failed to store updated taglist to S3", "error", err)
		return err
	}
	return nil
}

// GetTaglist gets Taglist
func (tls *taglistSync) GetCurrentTaglist() *sharedEntity.TagList {
	tls.mutex.Lock()
	defer tls.mutex.Unlock()

	copied := make([]sharedEntity.Tag, len(tls.tagList.Tags))
	copy(copied, tls.tagList.Tags)
	return &sharedEntity.TagList{Tags: copied}
}
