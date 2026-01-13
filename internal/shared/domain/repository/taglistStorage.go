package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	tracer         trace.Tracer
}

// NewTaglistStorage creates a new instance of TaglistRepository
func NewTaglistStorage(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.TagList],
	prefix string,
	tracer trace.Tracer,
) (TaglistStorage, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper, tracer); err != nil {
		return nil, err
	}

	return &taglistStorage{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		key:            "taglist.parquet",
		entryPrefix:    prefix,
		tracer:         tracer,
	}, nil
}

// CreateTaglist writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (tR *taglistStorage) CreateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx, taglist); err != nil {
		return err
	}

	ctx, span := tR.tracer.Start(ctx, "taglistStorage.CreateTaglist")
	defer span.End()
	span.SetAttributes(
		attribute.Int("taglist.tag_count", len(taglist.Tags)),
		attribute.String("taglist.key", tR.entryPrefix+tR.key),
	)

	if len(taglist.Tags) == 0 {
		err := fmt.Errorf("empty taglist")
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.ErrValidation
	}

	if err := validateTaglist(taglist); err != nil {
		err = fmt.Errorf("failed to validate TagList: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.ErrValidation
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(ctx, *taglist)
	if err != nil {
		err = fmt.Errorf("failed to write parquet: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize taglist")
		return errors.ErrInternalServer
	}

	tR.logger.Debug("TaglistStorage: parquet data created", slog.Int("size_bytes", len(parquetData)))

	var timestamp = fmt.Sprintf("%d", time.Now().Unix())
	metadata := map[string]string{
		"created": timestamp,
	}

	err = tR.s3Wrapper.UploadParquetFile(ctx, tR.entryPrefix+tR.key, parquetData, metadata)
	if err != nil {
		err = fmt.Errorf("failed to upload existing parquet: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload taglist parquet")
		return errors.ErrInternalServer
	}

	span.AddEvent("taglist stored", trace.WithAttributes(
		attribute.String("taglist.created_at", timestamp),
	))
	span.SetStatus(codes.Ok, "")
	return nil
}

// ReadTaglist retrieves the taglist from the database, downloading the Parquet file and reading its content.
func (tR *taglistStorage) ReadTaglist(ctx context.Context) (*entity.TagList, error) {
	if err := assert.NotNil(ctx); err != nil {
		tR.logger.Error(fmt.Sprintf("context cannot be nil, %s", err))
		return nil, errors.ErrInternalServer
	}

	ctx, span := tR.tracer.Start(ctx, "taglistStorage.ReadTaglist")
	defer span.End()
	span.SetAttributes(attribute.String("taglist.key", tR.entryPrefix+tR.key))

	parquetData, metadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, tR.entryPrefix+tR.key)
	if err != nil {
		err := fmt.Errorf("failed to download existing parquet: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download taglist parquet")
		return nil, errors.ErrInternalServer
	}
	tR.logger.Debug("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	taglists, err := tR.parquetWrapper.ReadStructsFromParquet(ctx, parquetData)
	if err != nil {
		err := fmt.Errorf("failed to read parquet data: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to parse taglist parquet")
		return nil, errors.ErrInternalServer
	}
	tR.logger.Debug("events read from parquet", slog.Int("count", len(taglists)))

	if len(taglists) == 0 {
		err := fmt.Errorf("no taglist found with key: %s", tR.entryPrefix+tR.key)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "no taglist found")
		return nil, errors.ErrInternalServer
	}

	firstEntry := taglists[0]

	if err := validateTaglist(&firstEntry); err != nil {
		err := fmt.Errorf("read invalid taglist: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "taglist validation failed")
		return nil, errors.ErrInternalServer
	}

	span.SetStatus(codes.Ok, "")
	return &firstEntry, nil
}

// UpdateTaglist updates the existing taglist, overwriting it's contents while keeping metadata.
func (tR *taglistStorage) UpdateTaglist(ctx context.Context, taglist *entity.TagList) error {
	if err := assert.NotNil(ctx, taglist); err != nil {
		return err
	}

	ctx, span := tR.tracer.Start(ctx, "taglistStorage.UpdateTaglist")
	defer span.End()
	span.SetAttributes(
		attribute.String("taglist.key", tR.entryPrefix+tR.key),
		attribute.Int("taglist.tag_count", len(taglist.Tags)),
	)

	if len(taglist.Tags) == 0 {
		err := fmt.Errorf("empty taglist")
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.ErrInternalServer
	}

	if err := validateTaglist(taglist); err != nil {
		err := fmt.Errorf("failed to validate TagList: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return errors.ErrValidation
	}

	_, oldmetadata, err := tR.s3Wrapper.DownloadParquetFile(ctx, tR.entryPrefix+tR.key)
	if err != nil {
		err = fmt.Errorf("failed to download data: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download existing taglist")
		return errors.ErrInternalServer
	}

	parquetData, err := tR.parquetWrapper.WriteStructToParquet(ctx, *taglist)
	if err != nil {
		err = fmt.Errorf("failed to write parquet: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize taglist")
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
		err = fmt.Errorf("failed to upload file: %w", err)
		tR.logger.Error((err.Error()))
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload updated taglist")
		return errors.ErrInternalServer
	}

	span.AddEvent("taglist updated", trace.WithAttributes(
		attribute.String("taglist.updated_at", timestamp),
	))
	span.SetStatus(codes.Ok, "")
	return nil
}

// TaglistExists checks whether an taglist is already stored
func (tR *taglistStorage) TaglistExists(ctx context.Context) (bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		return false, err
	}

	ctx, span := tR.tracer.Start(ctx, "taglistStorage.TaglistExists")
	defer span.End()
	span.SetAttributes(attribute.String("taglist.key", tR.entryPrefix+tR.key))

	exists, err := tR.s3Wrapper.FileExists(ctx, tR.entryPrefix+tR.key)
	if err != nil {
		tR.logger.Error("failed to check taglist existence", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check taglist existence")
		return false, errors.ErrInternalServer
	}

	span.SetAttributes(attribute.Bool("taglist.exists", exists))
	span.SetStatus(codes.Ok, "")
	return exists, nil
}

// validateTaglist validates the taglist before processing it
func validateTaglist(taglist *entity.TagList) error {
	if err := assert.NotNil(taglist); err != nil {
		return err
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
