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

// TestCaseStorageRepository defines the interface for a repository managing TestCase entities.
type TestCaseStorageRepository interface {
	// Create stores a new TestCase object in the underlying storage system.
	Create(ctx context.Context, obj *entity.TestCase) (string, error)

	// Read retrieves a TestCase object from storage by its key.
	Read(ctx context.Context, key string) (*entity.TestCase, error)

	// Update modifies an existing TestCase object in the storage system.
	Update(ctx context.Context, obj *entity.TestCase, key string) error

	// Delete removes a TestCase object from the storage system by its key.
	Delete(ctx context.Context, key string) error
}

// testCaseStorageRepository provides a repository implementation for TestCase entities,
// encapsulating logic for S3 and Parquet operations.
type testCaseStorageRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.TestCase]
	logger         *slog.Logger
}

// NewTestCaseStorageRepository creates a new repository for TestCase entities.
// It initializes the required S3 and Parquet wrappers.
// Returns the repository or an error.
func NewTestCaseStorageRepository(logger *slog.Logger) (TestCaseStorageRepository, error) {
	s3Config := wrapperEntity.S3Config{
		Bucket:    "autotester",
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}

	s3Wrapper, err := service.NewS3Wrapper(logger, s3Config)
	if err != nil {
		return nil, err
	}

	parquetWrapper, err := service.NewParquetWrapper[entity.TestCase](logger, service.DefaultParquetConfig())
	if err != nil {
		return nil, err
	}

	return &testCaseStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given TestCase object to Parquet format and uploads it to S3.
// Returns the generated S3 key or an error.
// nolint:dupl
func (r *testCaseStorageRepository) Create(ctx context.Context, obj *entity.TestCase) (string, error) {
	if err := testCaseValidationFunc(obj); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		r.logger.Error("create: writing struct to parquet failed",
			slog.String("type", "testcase"),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	key := generateTestCaseKey()
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
		slog.String("type", "testcase"),
	)

	return key, nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first TestCase found.
// Returns the TestCase or an error.
// nolint:dupl
func (r *testCaseStorageRepository) Read(ctx context.Context, key string) (*entity.TestCase, error) {
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
	if err := testCaseValidationFunc(&items[0]); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided TestCase.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *testCaseStorageRepository) Update(ctx context.Context, obj *entity.TestCase, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := testCaseValidationFunc(obj); err != nil {
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
			slog.String("type", "testcase"),
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
	)

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *testCaseStorageRepository) Delete(ctx context.Context, key string) error {
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

// generateTestCaseKey creates a unique S3 key for a TestCase object.
// The format is: "testCase/testCase_<timestamp>"
func generateTestCaseKey() string {
	timestamp := time.Now().Format("20060102150405000")
	return fmt.Sprintf("%s/%s_%s", "testCase", "testCase", timestamp)
}

// testCaseValidationFunc checks if a TestCase object is valid.
// Returns an error if any required field is empty or invalid.
func testCaseValidationFunc(testcase *entity.TestCase) error {
	if err := assert.NotNil(testcase); err != nil {
		return fmt.Errorf("obj must not be nil: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestID); err != nil {
		return fmt.Errorf("testcase.TestID must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.Description); err != nil {
		return fmt.Errorf("testcase.Description must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(string(testcase.Status)); err != nil {
		return fmt.Errorf("testcase.Status must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestCode.Code); err != nil {
		return fmt.Errorf("testcase.TestCode.Code must not be empty: %w", err)
	}
	return nil
}
