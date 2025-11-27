package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseStorageRepository defines the interface for a repository managing TestCase entities.
type TestcaseStorageRepository interface {
	// Create stores a new TestCase object in the underlying storage system.
	Create(ctx context.Context, obj *entity.TestCase, userId string) (string, error)

	// Read retrieves a TestCase object from storage by its key.
	Read(ctx context.Context, key string) (*entity.TestCase, error)

	// Update modifies an existing TestCase object in the storage system.
	Update(ctx context.Context, obj *entity.TestCase, key string) error

	// Delete removes a TestCase object from the storage system by its key.
	Delete(ctx context.Context, key string) error
}

// testcaseStorageRepository provides a repository implementation for TestCase entities,
// encapsulating logic for S3 and Parquet operations.
type testcaseStorageRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.TestCase]
	logger         *slog.Logger
	s3Prefix       string
}

// NewTestcaseStorageRepository creates a new repository for TestCase entities.
// Returns the repository or an error.
func NewTestcaseStorageRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.TestCase],
	s3Prefix string,
) (TestcaseStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper); err != nil {
		return nil, err
	}

	if err := assert.StringNotEmpty(s3Prefix); err != nil {
		return nil, err
	}

	return &testcaseStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given TestCase object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *testcaseStorageRepository) Create(ctx context.Context, obj *entity.TestCase, userId string) (string, error) {
	if err := validateTestCaseData(obj); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}
	if err := assert.StringNotEmpty(userId); err != nil {
		return "", fmt.Errorf("userId must not be empty")
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		return "", err
	}
	key := r.generateTestCaseKey(obj.TestID)
	metadata := map[string]string{
		"created": fmt.Sprintf("%d", time.Now().UTC().Unix()),
		"author":  userId,
	}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return "", err
	}

	r.logger.Debug("create: object successfully written and uploaded",
		slog.String("key", key),
		slog.String("type", "testcase"),
	)

	return key, nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first TestCase found.
// Returns the TestCase or an error.
// nolint:dupl
func (r *testcaseStorageRepository) Read(ctx context.Context, key string) (*entity.TestCase, error) {
	if err := assert.StringNotEmpty(key); err != nil {
		return nil, fmt.Errorf("key must not be empty")
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
		return nil, fmt.Errorf("no data found for key %s", key)
	}
	if err := validateTestCaseData(&items[0]); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the provided TestCase.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *testcaseStorageRepository) Update(ctx context.Context, obj *entity.TestCase, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := validateTestCaseData(obj); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	exists, err := r.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if key exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("cannot update: key does not exist")
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	metadata := map[string]string{}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload updated object: %w", err)
	}

	r.logger.Debug("update: object successfully overwritten",
		slog.String("key", key),
	)

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *testcaseStorageRepository) Delete(ctx context.Context, key string) error {
	if err := assert.StringNotEmpty(key); err != nil {
		return fmt.Errorf("key must not be empty")
	}

	if err := r.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		return err
	}

	r.logger.Debug("delete: object successfully deleted",
		slog.String("key", key),
	)
	return nil
}

// generateTestCaseKey creates a unique S3 key for a TestCase object.
// The format is: "testcase/<testCaseID>_<timestamp>.parquet"
func (r *testcaseStorageRepository) generateTestCaseKey(testcaseId string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s/%s_%d.parquet", r.s3Prefix, testcaseId, timestamp)
}

// validateTestCaseData checks if a TestCase object is valid.
// Returns an error if any required field is empty or invalid.
func validateTestCaseData(testcase *entity.TestCase) error {
	if err := assert.NotNil(testcase); err != nil {
		return fmt.Errorf("obj must not be nil: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestID); err != nil {
		return fmt.Errorf("testcase.TestID must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(string(testcase.Status)); err != nil {
		return fmt.Errorf("testcase.Status must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestCode.Code); err != nil {
		return fmt.Errorf("testcase.TestCode.Code must not be empty: %w", err)
	}
	return nil
}
