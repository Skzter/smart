package service

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"sync"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistSync defines interface for syncing taglist.
type TaglistSync interface {
	// SyncTaglist syncs stored taglist.
	SyncTaglist(context.Context, []string) error
}

type taglistSync struct {
	logger         *slog.Logger
	taglistService service.TaglistStorage
	tagList        []string
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
func (tls *taglistSync) SyncTaglist(ctx context.Context, taglist []string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	if len(taglist) == 0 {
		return errors.New("empty taglist")
	}

	tls.mutex.Lock()
	defer tls.mutex.Unlock()

	lenghtList := len(tls.tagList)
	for _, tag := range taglist {
		if !slices.Contains(tls.tagList, tag) {
			tls.tagList = append(tls.tagList, tag)
		}
	}

	if lenghtList != len(tls.tagList) {
		tls.logger.Info("UPDATES")
		err := tls.taglistService.StoreTaglist(ctx, tls.tagList)
		if err != nil {
			return err
		}
	}
	return nil
}
