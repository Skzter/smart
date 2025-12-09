package repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
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
	tracer         trace.Tracer
	entryPrefix    string
}

// NewDatabaseRepository creates a new instance of DatabaseRepository
func NewDatabaseRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry],
	tracer trace.Tracer,
	prefix string,
) (DatabaseRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper, tracer); err != nil {
		return nil, err
	}

	return &databaseRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		tracer:         tracer,
		entryPrefix:    prefix,
	}, nil
}

// CreateRequest writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (dbR *databaseRepository) CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		dbR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return sharedErrors.ErrInternalServer
	}

	ctx, span := dbR.tracer.Start(ctx, "databaseRepository.CreateRequest")
	defer span.End()

	if err := validateDbEntry(dbEntry); err != nil {
		dbR.logger.Error("dbEntry validation failed", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "dbEntry validation failed")
		return sharedErrors.ErrValidation
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(ctx, dbEntry)
	if err != nil {
		dbR.logger.Error("failed to write parquet", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write parquet")
		return sharedErrors.ErrGeneration
	}

	dbR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	var timestamp = fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": timestamp,
	}

	key := generateKey(dbEntry.Tags, timestamp)

	err = dbR.s3Wrapper.UploadParquetFile(ctx, dbR.entryPrefix+key, parquetData, metadata)
	if err != nil {
		dbR.logger.Error("failed to upload parquet file", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload parquet file")
		return sharedErrors.ErrGeneration
	}

	span.AddEvent("object successfully written and uploaded", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "databaseEntry"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// ReadRequest retrieves a request from the database by its key, downloading the Parquet file and reading its content.
func (dbR *databaseRepository) ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error) {
	if err := assert.NotNil(ctx); err != nil {
		dbR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return nil, sharedErrors.ErrGeneration
	}
	ctx, span := dbR.tracer.Start(ctx, "databaseRepository.ReadRequest")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		dbR.logger.Error(("key must not be empty"), slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return nil, sharedErrors.ErrValidation
	}

	parquetData, metadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		dbR.logger.Error("failed to download existing parquet", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download parquet file")
		return nil, sharedErrors.ErrInternalServer
	}
	dbR.logger.Debug("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	dbEntries, err := dbR.parquetWrapper.ReadStructsFromParquet(ctx, parquetData)
	if err != nil {
		dbR.logger.Error("failed to read parquet data", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read parquet data")
		return nil, sharedErrors.ErrInternalServer
	}
	dbR.logger.Debug("events read from parquet", slog.Int("count", len(dbEntries)))
	firstEntry := dbEntries[0]

	span.AddEvent("object successfully read", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "databaseEntry"),
	))
	span.SetStatus(codes.Ok, "")

	return &firstEntry, nil
}

// UpdateRequest updates an existing request in the database by downloading the Parquet file, modifying its content, and re-uploading it.
func (dbR *databaseRepository) UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error {
	if err := assert.NotNil(ctx); err != nil {
		dbR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return sharedErrors.ErrInternalServer
	}

	ctx, span := dbR.tracer.Start(ctx, "databaseRepository.UpdateRequest")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		dbR.logger.Error("key must not be empty", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return sharedErrors.ErrValidation
	}

	if err := validateDbEntry(dbEntry); err != nil {
		dbR.logger.Error("dbEntry validation failed", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "dbEntry validation failed")
		return sharedErrors.ErrValidation
	}

	_, oldmetadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		dbR.logger.Error("failed to download existing parquet", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download parquet file")
		return sharedErrors.ErrInternalServer
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(ctx, dbEntry)
	if err != nil {
		dbR.logger.Error("failed to write parquet", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write parquet")
		return sharedErrors.ErrInternalServer
	}

	dbR.logger.Debug("parquet data created", slog.Int("size_bytes", len(parquetData)))

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": oldmetadata["created"],
		"updated": timestamp,
	}

	err = dbR.s3Wrapper.UploadParquetFile(ctx, dbR.entryPrefix+key, parquetData, metadata)
	if err != nil {
		dbR.logger.Error("failed to upload parquet file", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload parquet file")
		return sharedErrors.ErrInternalServer
	}

	span.AddEvent("object successfully overwritten", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "databaseEntry"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// DeleteRequest deletes a request from the database by removing the Parquet file associated with the given key.
func (dbR *databaseRepository) DeleteRequest(ctx context.Context, key string) error {
	if err := assert.NotNil(ctx); err != nil {
		dbR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return sharedErrors.ErrInternalServer
	}

	ctx, span := dbR.tracer.Start(ctx, "databaseRepository.DeleteRequest")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		dbR.logger.Error("key must not be empty", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return sharedErrors.ErrValidation
	}

	err := dbR.s3Wrapper.DeleteParquetFile(ctx, dbR.entryPrefix+key)
	if err != nil {
		dbR.logger.Error("failed to delete parquet file", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete parquet file")
		return sharedErrors.ErrInternalServer
	}

	dbR.logger.Debug("file deleted successfully", slog.String("key", key))

	span.AddEvent("object successfully deleted", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "databaseEntry"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

func (dbR *databaseRepository) ListAllKeys(ctx context.Context) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		dbR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return nil, sharedErrors.ErrInternalServer
	}

	ctx, span := dbR.tracer.Start(ctx, "databaseRepository.ListAllKeys")
	defer span.End()

	keys, err := dbR.s3Wrapper.ListParquetFiles(ctx, dbR.entryPrefix)
	if err != nil {
		dbR.logger.Error("failed to list parquet files", slog.String("error", err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list parquet files")
		return nil, sharedErrors.ErrInternalServer
	}

	span.AddEvent("ListAllKeys: finished loading keys", trace.WithAttributes(
		attribute.String("type", "databaseEntry"),
		attribute.Int("key_count", len(keys)),
	))
	span.SetStatus(codes.Ok, "")

	return keys, nil
}

// validateDbEntry validates the database entry before processing it
func validateDbEntry(dbEntry entity.DatabaseEntry) error {
	if err := validateRequest(dbEntry.Request); err != nil {
		return fmt.Errorf("failed validate request: %w", err)
	}
	if err := validateResponse(dbEntry.Response); err != nil {
		return fmt.Errorf("failed validate response: %w", err)
	}
	if err := validateTags(dbEntry.Tags); err != nil {
		return fmt.Errorf("failed validate tags: %w", err)
	}
	return nil
}

// validateRequest validates the request part of the database entry
func validateRequest(rq entity.Request) error {
	if len(rq.Header) == 0 {
		return fmt.Errorf("header must not be empty")
	}

	if err := assert.StringNotEmpty(rq.Tags); err != nil {
		return fmt.Errorf("tags must not be empty: %w", err)
	}

	if err := assert.StringNotEmpty(rq.Destination); err != nil {
		return fmt.Errorf("destination must not be empty: %w", err)
	}

	if err := assert.StringNotEmpty(rq.Body); err != nil {
		return fmt.Errorf("body must not be empty: %w", err)
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
func validateTags(t *sharedEntity.TagList) error {
	if t == nil || len(t.Tags) == 0 {
		return fmt.Errorf("tags must not be empty")
	}
	return nil
}

// generateKey creates a unique key for the database entry based on its tags and a Unix timestamp
func generateKey(taglist *sharedEntity.TagList, unixTimestamp string) string {
	tags := make([]string, len(taglist.Tags))
	for i, tag := range taglist.Tags {
		tags[i] = strings.ToLower(strings.TrimSpace(tag.Name))
	}
	return fmt.Sprintf("%s-%s", strings.Join(tags, "-"), unixTimestamp)
}
