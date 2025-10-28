package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// EntryPrefix is the prefix used for database entries in the S3 storage.
const EntryPrefix = "taglist/"

// TaglistRepository defines the interface for database operations related to supplier data.
type TaglistRepository interface {
	// creates a request into the database
	CreateTaglist(ctx context.Context, TagList entity.TagListEntity) error
	// gets Request (access through key)
	ReadTaglist(ctx context.Context) (*entity.TagListEntity, error)
	// changes the content of a request (access through key)
	UpdateTaglist(ctx context.Context, TagList entity.TagListEntity) error
	// checks if a checklist exists in s3
	TaglistExists(ctx context.Context) (bool, error)
}

type taglistRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.TagListEntity]
	logger         *slog.Logger
	key            string
}

// NewTaglistRepository creates a new instance of DatabaseRepository
func NewTaglistRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.TagListEntity],
) (TaglistRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}

	return &taglistRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		key:            "taglist",
	}, nil
}

// CreateRequest writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (tR *taglistRepository) CreateTaglist(ctx context.Context, taglist entity.TagListEntity) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := validateTaglist(taglist); err != nil {
		return fmt.Errorf("failed to validate TagList: %w", err)
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(taglist)
	if err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	tR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	var timestamp = fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": timestamp,
	}

	err = tR.s3Wrapper.UploadParquetFile(ctx, EntryPrefix+tR.key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload existing parquet: %w", err)
	}

	return nil
}

// ReadRequest retrieves a request from the database by its key, downloading the Parquet file and reading its content.
func (tR *taglistRepository) ReadTaglist(ctx context.Context) (*entity.TagListEntity, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(tR.key); err != nil {
		return nil, fmt.Errorf("key must not be empty: %w", err)
	}

	parquetData, metadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, EntryPrefix+tR.key)
	if err != nil {
		return nil, fmt.Errorf("failed to download existing parquet: %w", err)
	}
	tR.logger.Debug("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	dbEntries, err := tR.parquetWrapper.ReadStructsFromParquet(parquetData)
	if err != nil {
		return nil, fmt.Errorf("failed to read parquet data: %w", err)
	}
	tR.logger.Debug("events read from parquet", slog.Int("count", len(dbEntries)))
	firstEntry := dbEntries[0]

	if err := validateTaglist(firstEntry); err != nil {
		return nil, fmt.Errorf("read invalid taglist: %w", err)
	}

	return &firstEntry, nil
}

// UpdateRequest updates an existing request in the database by downloading the Parquet file, modifying its content, and re-uploading it.
func (tR *taglistRepository) UpdateTaglist(ctx context.Context, taglist entity.TagListEntity) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(tR.key); err != nil {
		return fmt.Errorf("key must not be empty: %w", err)
	}

	if err := validateTaglist(taglist); err != nil {
		return fmt.Errorf("failed to validate TagList: %w", err)
	}

	_, oldmetadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, EntryPrefix+tR.key)
	if err != nil {
		return fmt.Errorf("failed to download data: %w", err)
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(taglist)
	if err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	tR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": oldmetadata["created"],
		"updated": timestamp,
	}

	err = tR.s3Wrapper.UploadParquetFile(ctx, EntryPrefix+tR.key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (tR *taglistRepository) TaglistExists(ctx context.Context) (bool, error) {
	return tR.s3Wrapper.FileExists(ctx, EntryPrefix+tR.key)
}

// validateTaglist validates the database entry before processing it
func validateTaglist(taglist entity.TagListEntity) error {
	if len(taglist.Tags) == 0 {
		return fmt.Errorf("taglist is empty: %s", taglist)
	}
	for _, tag := range taglist.Tags {
		if err := assert.StringNotEmpty(tag); err != nil {
			return err
		}
	}

	return nil
}
