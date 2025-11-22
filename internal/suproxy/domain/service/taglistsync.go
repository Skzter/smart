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
		logger.Error("Failed to get initial taglist", "error", err)
		return nil, err
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
	if len(taglist.Tags) == 0 {
		return errors.New("empty taglist")
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

	for _, tag := range onlineTaglist.Tags {
		if !slices.Contains(tls.tagList.Tags, tag) {
			tls.tagList.Tags = append(tls.tagList.Tags, tag)
		}
	}

	allExist = true
	for _, t := range incomingTags {
		if !slices.Contains(tls.tagList.Tags, t) {
			allExist = false
			break
		}
	}
	if allExist {
		return nil
	}

	// 5) fehlende incomingTags hinzufügen und upload zu S3
	for _, t := range incomingTags {
		if !slices.Contains(tls.tagList.Tags, t) {
			tls.tagList.Tags = append(tls.tagList.Tags, t)
		}
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
