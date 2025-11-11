package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// DatabaseRepository defines the interface for database operations related to supplier data.
type DatabaseRepository interface {
	// creates a request into the database
	CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error
	// gets Request (access through key)
	ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error)
	// changes the content of a request (access through key)
	UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error
	// deletes a request from the database (access through key)
	DeleteRequest(ctx context.Context, key string) error
	// ListKeysFromFile lists all keys from the database
	ListAllKeys(ctx context.Context) ([]string, error)
}

type databaseRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry]
	logger         *slog.Logger
	entryPrefix    string
}

// NewDatabaseRepository creates a new instance of DatabaseRepository
func NewDatabaseRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry],
	prefix string,
) (DatabaseRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}

	return &databaseRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		entryPrefix:    prefix,
	}, nil
}

// CreateRequest writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (dbR *databaseRepository) CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := validateDbEntry(dbEntry); err != nil {
		return fmt.Errorf("failed to validate dbEntry: %w", err)
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(dbEntry)
	if err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	dbR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	var timestamp = fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": timestamp,
	}

	key := generateKey(dbEntry.Tags, timestamp)

	err = dbR.s3Wrapper.UploadParquetFile(ctx, dbR.entryPrefix+key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload existing parquet: %w", err)
	}

	return nil
}

// ReadRequest retrieves a request from the database by its key, downloading the Parquet file and reading its content.
func (dbR *databaseRepository) ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return nil, fmt.Errorf("key must not be empty: %w", err)
	}

	parquetData, metadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		return nil, fmt.Errorf("failed to download existing parquet: %w", err)
	}
	dbR.logger.Debug("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	dbEntries, err := dbR.parquetWrapper.ReadStructsFromParquet(parquetData)
	if err != nil {
		return nil, fmt.Errorf("failed to read parquet data: %w", err)
	}
	dbR.logger.Debug("events read from parquet", slog.Int("count", len(dbEntries)))
	firstEntry := dbEntries[0]
	return &firstEntry, nil
}

// UpdateRequest updates an existing request in the database by downloading the Parquet file, modifying its content, and re-uploading it.
func (dbR *databaseRepository) UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty: %w", err)
	}

	if err := validateDbEntry(dbEntry); err != nil {
		return fmt.Errorf("failed to validate dbEntry: %w", err)
	}

	_, oldmetadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		return fmt.Errorf("failed to download data: %w", err)
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(dbEntry)
	if err != nil {
		return fmt.Errorf("failed to write parquet: %w", err)
	}

	dbR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": oldmetadata["created"],
		"updated": timestamp,
	}

	err = dbR.s3Wrapper.UploadParquetFile(ctx, dbR.entryPrefix+key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DeleteRequest deletes a request from the database by removing the Parquet file associated with the given key.
func (dbR *databaseRepository) DeleteRequest(ctx context.Context, key string) error {
	if err := assert.NotNil(ctx); err != nil {
		return fmt.Errorf("context cannot be nil, %w", err)
	}
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty: %w", err)
	}

	err := dbR.s3Wrapper.DeleteParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	dbR.logger.Debug("file deleted successfully", slog.String("key", key))

	return nil
}

func (dbR *databaseRepository) ListAllKeys(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, fmt.Errorf("context cannot be nil, %w", err)
	}
	keys, err := dbR.s3Wrapper.ListParquetFiles(ctx, dbR.entryPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list parquet files: %w", err)
	}

	return keys, err
}

// validateDbEntry validates the database entry before processing it
func validateDbEntry(dbEntry entity.DatabaseEntry) error {
	if err := assert.StringNotEmpty(dbEntry.Request); err != nil {
		return fmt.Errorf("request must not be empty: %w", err)
	}
	if err := validateResponse(dbEntry.Response); err != nil {
		return fmt.Errorf("failed to validate response: %w", err)
	}
	if err := validateTags(dbEntry.Tags); err != nil {
		return fmt.Errorf("failed to validate tags: %w", err)
	}
	return nil
}

// validateResponse validates the response part of the database entry
func validateResponse(rp entity.Response) error {
	if err := assert.StringNotEmpty(rp.Response); err != nil {
		return fmt.Errorf("response must not be empty: %w", err)
	}
	return nil
}

// validateTags validates the tags associated with the database entry
func validateTags(t []string) error {
	if len(t) == 0 {
		return fmt.Errorf("tags must not be empty")
	}
	return nil
}

// generateKey creates a unique key for the database entry based on its tags and a Unix timestamp
func generateKey(tags []string, unixTimestamp string) string {
	return fmt.Sprintf("%s-%s", strings.Join(tags, "-"), unixTimestamp)
}
