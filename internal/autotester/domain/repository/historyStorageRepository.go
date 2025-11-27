package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
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
}

// NewSessionSummaryStorageRepository creates a new repository for SessionSummary entities.
// Returns the repository or an error.
// nolint:lll
func NewSessionSummaryStorageRepository(logger *slog.Logger, s3Wrapper service.S3StorageWrapper, parquetWrapper service.ParquetFileWrapper[entity.SessionSummary]) (SessionSummaryStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}
	return &sessionSummaryStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given SessionSummary object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Create(ctx context.Context, obj *entity.SessionSummary) error {
	if err := validateHistoryData(obj); err != nil {
		r.logger.Error(fmt.Sprintf("validation failed: %s", err))
		return errors.ErrValidation
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to write parquet: %s", err))
		return errors.ErrInternalServer
	}

	key := generateSessionSummaryKey()
	metadata := map[string]string{
		"created": fmt.Sprintf("%d", time.Now().UTC().Unix()),
	}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return err
	}

	r.logger.Debug("create: object successfully written and uploaded",
		slog.String("key", key),
		slog.String("type", "sessionSummary"),
	)

	return nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first SessionSummary found.
// Returns the SessionSummary or an error.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Read(ctx context.Context, key string) (*entity.SessionSummary, error) {
	if err := assert.StringNotEmpty(key); err != nil {
		r.logger.Error(fmt.Sprintf("key must not be empty %s", err))
		return nil, errors.ErrValidation
	}

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		return nil, err
	}

	items, err := r.parquetWrapper.ReadStructsFromParquet(data)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		r.logger.Error(fmt.Sprintf("no data found for key %s", key))
		return nil, errors.ErrInternalServer
	}
	if err := validateHistoryData(&items[0]); err != nil {
		r.logger.Error(fmt.Sprintf("validation failed: %s", err))
		return nil, errors.ErrInternalServer
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided SessionSummary.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Update(ctx context.Context, obj *entity.SessionSummary, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		r.logger.Error(fmt.Sprintf("key must not be empty: %s", err))
		return errors.ErrValidation
	}

	if err := validateHistoryData(obj); err != nil {
		r.logger.Error(fmt.Sprintf("validation failed: %s", err))
		return errors.ErrValidation
	}

	exists, err := r.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to check if key exists: %s", err))
		return errors.ErrInternalServer
	}
	if !exists {
		r.logger.Error(fmt.Sprintf("cannot update: key does not exist %s", key))
		return errors.ErrGeneration
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to serialize object: %s", err))
		return errors.ErrInternalServer
	}

	metadata := map[string]string{}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to upload updated object: %s", err))
		return errors.ErrInternalServer
	}

	r.logger.Debug("update: object successfully overwritten",
		slog.String("key", key),
		slog.String("type", "sessionSummary"),
	)

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Delete(ctx context.Context, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		r.logger.Error(fmt.Sprintf("key must not be empty: %s", err))
		return errors.ErrValidation
	}

	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		r.logger.Error(fmt.Sprintf("failed to delete object: %s", err))
		return errors.ErrInternalServer
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", key),
	)
	return nil
}

func (r *sessionSummaryStorageRepository) ListAll(ctx context.Context) ([]entity.SessionSummary, error) {
	result := make([]entity.SessionSummary, 0)
	keys, err := r.s3Wrapper.ListParquetFiles(ctx, fmt.Sprint(prefixSessionSummary+"/"))
	if err != nil {
		r.logger.Error(fmt.Sprintf("failed to list session summary parquet files: %s", err))
		return nil, errors.ErrInternalServer
	}
	if len(keys) == 0 {
		r.logger.Error(fmt.Sprintf("no session summary files found in storage: %s", err))
		return nil, errors.ErrInternalServer
	}

	for _, key := range keys {
		data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
		if err != nil {
			r.logger.Warn("ListAll: failed to download parquet file",
				slog.String("key", key),
				slog.String("error", err.Error()))
			continue
		}
		items, err := r.parquetWrapper.ReadStructsFromParquet(data)
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

	r.logger.Debug("ListAll: finished loading session summaries", slog.Int("count", len(result)))
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
