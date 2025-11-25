package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TaglistStorage defines the interface for database operations on the taglist.
type TaglistStorage interface {
	// creates taglist in database
	CreateTaglist(ctx context.Context, TagList *entity.TagList) error
	// retrieves taglist from database
	ReadTaglist(ctx context.Context) (*entity.TagList, error)
	// overwrites the stored taglist with the one provided
	UpdateTaglist(ctx context.Context, TagList *entity.TagList) error
	// checks if a taglist exists in s3
	TaglistExists(ctx context.Context) (bool, error)
}

type taglistStorage struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.TagList]
	logger         *slog.Logger
	key            string
	entryPrefix    string
}

// NewTaglistStorage creates a new instance of TaglistRepository
func NewTaglistStorage(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.TagList],
	prefix string,
) (TaglistStorage, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}

	return &taglistStorage{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		key:            "taglist.parquet",
		entryPrefix:    prefix,
	}, nil
}

// CreateTaglist writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (tR *taglistStorage) CreateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx, taglist); err != nil {
		return err
	}
	if err := validateTaglist(taglist); err != nil {
		tR.logger.Error(fmt.Sprintf("validation failed, for TagList: %s", err))
		return errors.ErrValidation
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(*taglist)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to write parquet: %s", err))
		return errors.ErrInternalServer
	}

	tR.logger.Debug("TaglistStorage: parquet data created", slog.Int("size_bytes", len(parquetData)))

	var timestamp = fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": timestamp,
	}

	err = tR.s3Wrapper.UploadParquetFile(ctx, tR.entryPrefix+tR.key, parquetData, metadata)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to upload existing parquet: %s", err))
		return errors.ErrInternalServer
	}

	return nil
}

// ReadTaglist retrieves the taglist from the database, downloading the Parquet file and reading its content.
func (tR *taglistStorage) ReadTaglist(ctx context.Context) (*entity.TagList, error) {
	if err := assert.NotNil(ctx); err != nil {
		tR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return nil, errors.ErrInternalServer
	}

	parquetData, metadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, tR.entryPrefix+tR.key)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to download existing parquet: %s", err))
		return nil, errors.ErrInternalServer
	}
	tR.logger.Debug("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	taglists, err := tR.parquetWrapper.ReadStructsFromParquet(parquetData)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to read parquet data: %s", err))
		return nil, errors.ErrInternalServer
	}
	tR.logger.Debug("events read from parquet", slog.Int("count", len(taglists)))

	if len(taglists) == 0 {
		tR.logger.Error(fmt.Sprintf("no taglist found with key: %s", tR.entryPrefix+tR.key))
		return nil, errors.ErrInternalServer
	}

	firstEntry := taglists[0]

	if err := validateTaglist(&firstEntry); err != nil {
		tR.logger.Error(fmt.Sprintf("read invalid taglist: %s", err))
		return nil, errors.ErrInternalServer
	}

	return &firstEntry, nil
}

// UpdateTaglist updates the existing taglist, overwriting it's contents while keeping metadata.
func (tR *taglistStorage) UpdateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx, taglist); err != nil {
		return err
	}

	if err := validateTaglist(taglist); err != nil {
		tR.logger.Error(fmt.Sprintf("failed to validate TagList: %s", err))
		return errors.ErrValidation
	}

	_, oldmetadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, tR.entryPrefix+tR.key)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to download data: %s", err))
		return errors.ErrInternalServer
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(*taglist)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to write parquet: %s", err))
		return errors.ErrInternalServer
	}

	tR.logger.Debug("TaglistStorage: parquet data created", slog.Int("size_bytes", len(parquetData)))

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": oldmetadata["created"],
		"updated": timestamp,
	}

	err = tR.s3Wrapper.UploadParquetFile(ctx, tR.entryPrefix+tR.key, parquetData, metadata)
	if err != nil {
		tR.logger.Error(fmt.Sprintf("failed to upload file: %s", err))
		return errors.ErrInternalServer
	}

	return nil
}

// TaglistExists checks whether an taglist is already stored
func (tR *taglistStorage) TaglistExists(ctx context.Context) (bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		return false, err
	}
	return tR.s3Wrapper.FileExists(ctx, tR.entryPrefix+tR.key)
}

// validateTaglist validates the taglist before processing it
func validateTaglist(taglist *entity.TagList) error {
	if err := assert.NotNil(taglist); err != nil {
		return err
	}

	if len(taglist.Tags) == 0 {
		return fmt.Errorf("taglist is empty: %s", taglist)
	}
	for _, tag := range taglist.Tags {
		if err := assert.StringNotEmpty(tag.Name); err != nil {
			return fmt.Errorf("tag name is empty: %s", tag.Name)
		}
		if err := assert.StringNotEmpty(tag.Description); err != nil {
			return fmt.Errorf("tag description is empty: %s", tag.Description)
		}
	}
	return nil
}
