package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// SessionSummaryStorageRepository defines the interface for a repository managing SessionSummary entities.
type SessionSummaryStorageRepository interface {
	// Create stores a new SessionSummary object in the underlying storage system.
	Create(ctx context.Context, obj *entity.SessionSummary) error

	// Read retrieves a SessionSummary object from storage by its key.
	Read(ctx context.Context, key string) (*entity.SessionSummary, error)

	// Update overwrites an existing SessionSummary object in storage.
	Update(ctx context.Context, obj *entity.SessionSummary, key string) error

	// Delete removes a SessionSummary object from storage by its key.
	Delete(ctx context.Context, key string) error

	ListAll(ctx context.Context) ([]entity.SessionSummary, error)
}

const prefixSessionSummary = "sessionSummary"

// sessionSummaryStorageRepository implements the SessionSummaryStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type sessionSummaryStorageRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.SessionSummary]
	logger         *slog.Logger
	tracer         trace.Tracer
}

// NewSessionSummaryStorageRepository creates a new repository for SessionSummary entities.
// Returns the repository or an error.
// nolint:lll
func NewSessionSummaryStorageRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.SessionSummary],
	tracer trace.Tracer,
) (SessionSummaryStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper, tracer); err != nil {
		return nil, err
	}

	return &sessionSummaryStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		tracer:         tracer,
	}, nil
}

// Create serializes the given SessionSummary object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Create(ctx context.Context, obj *entity.SessionSummary) error {
	ctx, span := r.tracer.Start(ctx, "sessionSummaryStorageRepository.Create")
	defer span.End()

	if err := validateHistoryData(obj); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return fmt.Errorf("validation failed: %w", err)
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write parquet")
		return err
	}

	key := generateSessionSummaryKey()
	metadata := map[string]string{
		"created": fmt.Sprintf("%d", time.Now().UTC().Unix()),
	}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload parquet file")
		return err
	}

	span.AddEvent("object successfully written and uploaded", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "sessionSummary"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first SessionSummary found.
// Returns the SessionSummary or an error.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Read(ctx context.Context, key string) (*entity.SessionSummary, error) {
	ctx, span := r.tracer.Start(ctx, "sessionSummaryStorageRepository.Read")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return nil, fmt.Errorf("key must not be empty")
	}

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to download parquet file")
		return nil, err
	}

	items, err := r.parquetWrapper.ReadStructsFromParquet(ctx, data)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read parquet data")
		return nil, err
	}
	if len(items) == 0 {
		err := fmt.Errorf("no data found for key %s", key)
		span.RecordError(err)
		span.SetStatus(codes.Error, "no data found")
		return nil, err
	}
	if err := validateHistoryData(&items[0]); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	span.AddEvent("object successfully read", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "sessionSummary"),
	))
	span.SetStatus(codes.Ok, "")

	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided SessionSummary.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Update(ctx context.Context, obj *entity.SessionSummary, key string) error {
	ctx, span := r.tracer.Start(ctx, "sessionSummaryStorageRepository.Update")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return fmt.Errorf("key must not be empty")
	}

	if err := validateHistoryData(obj); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return fmt.Errorf("validation failed: %w", err)
	}

	exists, err := r.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check if key exists")
		return fmt.Errorf("failed to check if key exists: %w", err)
	}
	if !exists {
		err := fmt.Errorf("cannot update: key does not exist")
		span.RecordError(err)
		span.SetStatus(codes.Error, "key does not exist")
		return err
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to serialize object")
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	metadata := map[string]string{}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload updated object")
		return fmt.Errorf("failed to upload updated object: %w", err)
	}

	span.AddEvent("object successfully overwritten", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "sessionSummary"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Delete(ctx context.Context, key string) error {
	ctx, span := r.tracer.Start(ctx, "sessionSummaryStorageRepository.Delete")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return fmt.Errorf("key must not be empty")
	}

	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete parquet file")
		return err
	}

	span.AddEvent("object successfully deleted", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "sessionSummary"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

func (r *sessionSummaryStorageRepository) ListAll(ctx context.Context) ([]entity.SessionSummary, error) {
	ctx, span := r.tracer.Start(ctx, "sessionSummaryStorageRepository.ListAll")
	defer span.End()

	result := make([]entity.SessionSummary, 0)
	keys, err := r.s3Wrapper.ListParquetFiles(ctx, fmt.Sprint(prefixSessionSummary+"/"))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list parquet files")
		return nil, fmt.Errorf("failed to list session summary parquet files: %w", err)
	}
	if len(keys) == 0 {
		err := fmt.Errorf("no session summary files found in storage")
		span.RecordError(err)
		span.SetStatus(codes.Error, "no files found")
		return nil, err
	}

	for _, key := range keys {
		data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Warn("ListAll: failed to download parquet file",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}
		items, err := r.parquetWrapper.ReadStructsFromParquet(ctx, data)
		if err != nil {
			r.logger.Warn("ListAll: failed to read parquet data",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}
		for _, item := range items {
			if err := validateHistoryData(&item); err == nil {
				result = append(result, item)
			} else {
				r.logger.Warn("ListAll: skipping invalid session summary",
					slog.String("key", key),
					slog.String("error", err.Error()))
			}
		}
	}

	span.AddEvent("ListAll: finished loading session summaries", trace.WithAttributes(
		attribute.String("type", "sessionSummary"),
		attribute.Int("count", len(result)),
	))
	span.SetStatus(codes.Ok, "")

	return result, nil
}

// generateSessionSummaryKey creates a unique S3 key for a SessionSummary object.
// The format is: "sessionSummary/sessionSummary_<timestamp>"
func generateSessionSummaryKey() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%s_%d", prefixSessionSummary, prefixSessionSummary, timestamp)
}

// validateHistoryData validates a SessionSummary entity.
// Returns an error if any required field is empty or invalid.
func validateHistoryData(summary *entity.SessionSummary) error {
	if err := assert.NotNil(summary); err != nil {
		return fmt.Errorf("obj must not be nil: %w", err)
	}
	if err := assert.StringNotEmpty(summary.Summary); err != nil {
		return fmt.Errorf("sessionSummary.Summary must not be empty: %w", err)
	}
	if summary.Messages == nil {
		return fmt.Errorf("sessionSummary.Messages must not be nil")
	}
	for i, msg := range summary.Messages {
		if msg == nil {
			return fmt.Errorf("sessionSummary.Messages[%d] must not be nil", i)
		}
	}
	return nil
}
