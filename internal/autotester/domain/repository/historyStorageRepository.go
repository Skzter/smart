package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// SessionSummaryStorageRepository defines the interface for a repository managing SessionSummary entities.
type SessionSummaryStorageRepository interface {
	// Create stores a new SessionSummary object in the underlying storage system.
	Create(ctx context.Context, obj *entity.SessionSummary) (string, error)

	// Read retrieves a SessionSummary object from storage by its key.
	Read(ctx context.Context, key string) (*entity.SessionSummary, error)

	// Update overwrites an existing SessionSummary object in storage.
	Update(ctx context.Context, obj *entity.SessionSummary, key string) error

	// Delete removes a SessionSummary object from storage by its key.
	Delete(ctx context.Context, key string) error

	ListAll(ctx context.Context) ([]entity.SessionSummary, error)
}

// sessionSummaryStorageRepository implements the SessionSummaryStorageRepository interface
// and encapsulates logic for S3 and Parquet operations.
type sessionSummaryStorageRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.SessionSummary]
	logger         *slog.Logger
}

// NewSessionSummaryStorageRepository creates a new repository for SessionSummary entities.
// It initializes the required S3 and Parquet wrappers.
// Returns the repository or an error.
func NewSessionSummaryStorageRepository(logger *slog.Logger) (SessionSummaryStorageRepository, error) {
	s3Config := wrapperEntity.S3Config{
		Bucket:    "autotester",
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}

	s3Wrapper, err := service.NewS3Wrapper(logger, s3Config)
	if err != nil {
		return nil, err
	}

	parquetWrapper, err := service.NewParquetWrapper[entity.SessionSummary](logger, service.DefaultParquetConfig())
	if err != nil {
		return nil, err
	}

	return &sessionSummaryStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given SessionSummary object to Parquet format and uploads it to S3.
// Returns the generated S3 key or an error.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Create(ctx context.Context, obj *entity.SessionSummary) (string, error) {
	if err := sessionSummaryValidationFunc(obj); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		r.logger.Error("create: writing struct to parquet failed",
			slog.String("type", "sessionSummary"),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	key := generateSessionSummaryKey()
	metadata := map[string]string{
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		r.logger.Error("create: uploading parquet file to S3 failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	r.logger.Info("create: object successfully written and uploaded",
		slog.String("key", key),
		slog.String("type", "sessionSummary"),
	)

	return key, nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first SessionSummary found.
// Returns the SessionSummary or an error.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Read(ctx context.Context, key string) (*entity.SessionSummary, error) {
	if err := assert.StringNotEmpty(key); err != nil {
		return nil, fmt.Errorf("key must not be empty")
	}

	data, _, err := r.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		r.logger.Error("read: downloading parquet failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	items, err := r.parquetWrapper.ReadStructsFromParquet(data)
	if err != nil {
		r.logger.Error("read: parsing parquet data failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no data found for key %s", key)
	}
	if err := sessionSummaryValidationFunc(&items[0]); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided SessionSummary.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *sessionSummaryStorageRepository) Update(ctx context.Context, obj *entity.SessionSummary, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := sessionSummaryValidationFunc(obj); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	exists, err := r.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		r.logger.Error("update: failed to check existence",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to check if key exists: %w", err)
	}
	if !exists {
		r.logger.Error("update: key does not exist, aborting",
			slog.String("key", key),
		)
		return fmt.Errorf("cannot update: key does not exist")
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		r.logger.Error("update: writing struct to parquet failed",
			slog.String("type", "sessionSummary"),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	metadata := map[string]string{}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload updated object: %w", err)
	}

	r.logger.Info("update: object successfully overwritten",
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
		return fmt.Errorf("key must not be empty")
	}

	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		r.logger.Error("delete: removing parquet failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return err
	}

	r.logger.Info("delete: object successfully deleted",
		slog.String("key", key),
	)
	return nil
}

func (r *sessionSummaryStorageRepository) ListAll(ctx context.Context) ([]entity.SessionSummary, error) {
	const prefix = "sessionSummary/"
	result := make([]entity.SessionSummary, 0)

	keys, err := r.s3Wrapper.ListParquetFiles(ctx, prefix)
	if err != nil {
		r.logger.Error("ListAll: failed to list parquet files", slog.String("error", err.Error()))
		return result, fmt.Errorf("failed to list session summary parquet files: %w", err)
	}
	if len(keys) == 0 {
		r.logger.Info("ListAll: no session summary files found")
		return result, fmt.Errorf("no session summary files found in storage")
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
			if err := sessionSummaryValidationFunc(&item); err == nil {
				result = append(result, item)
			} else {
				r.logger.Warn("ListAll: skipping invalid session summary",
					slog.String("key", key),
					slog.String("error", err.Error()))
			}
		}
	}

	r.logger.Info("ListAll: finished loading session summaries", slog.Int("count", len(result)))
	return result, nil
}

// generateSessionSummaryKey creates a unique S3 key for a SessionSummary object.
// The format is: "sessionSummary/sessionSummary_<timestamp>"
func generateSessionSummaryKey() string {
	timestamp := time.Now().Format("20060102150405000")
	return fmt.Sprintf("%s/%s_%s", "sessionSummary", "sessionSummary", timestamp)
}

// sessionSummaryValidationFunc validates a SessionSummary entity.
// Returns an error if any required field is empty or invalid.
func sessionSummaryValidationFunc(summary *entity.SessionSummary) error {
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
