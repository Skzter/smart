package repository

import (
	"context"
	"errors"
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

// TestcaseStorageRepository defines the interface for a repository managing TestCase entities.
type TestcaseStorageRepository interface {
	// Create stores a new TestCase object in the underlying storage system.
	Create(ctx context.Context, obj *entity.TestCase, userId string) (string, error)

	// Read retrieves a TestCase object from storage by its key.
	Read(ctx context.Context, key string) (*entity.TestCase, error)

	ReadAllMetadata(ctx context.Context) ([]*entity.TestcaseMetadata, error)

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
	tracer         trace.Tracer
}

// NewTestcaseStorageRepository creates a new repository for TestCase entities.
// Returns the repository or an error.
func NewTestcaseStorageRepository(
	logger *slog.Logger,
	s3Wrapper service.S3StorageWrapper,
	parquetWrapper service.ParquetFileWrapper[entity.TestCase],
	s3Prefix string,
	tracer trace.Tracer,
) (TestcaseStorageRepository, error) {
	if err := assert.NotNil(logger, s3Wrapper, parquetWrapper, tracer); err != nil {
		return nil, err
	}

	if err := assert.StringNotEmpty(s3Prefix); err != nil {
		return nil, err
	}

	return &testcaseStorageRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
		s3Prefix:       s3Prefix,
		tracer:         tracer,
	}, nil
}

// Create serializes the given TestCase object to Parquet format and uploads it to S3.
// nolint:dupl
func (r *testcaseStorageRepository) Create(ctx context.Context, obj *entity.TestCase, userId string) (string, error) {
	ctx, span := r.tracer.Start(ctx, "testCaseStorageRepository.Create")
	defer span.End()

	if err := validateTestCaseData(obj); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return "", fmt.Errorf("validation failed: %w", err)
	}
	if err := assert.StringNotEmpty(userId); err != nil {
		return "", fmt.Errorf("userId must not be empty")
	}

	parquetData, err := r.parquetWrapper.WriteStructToParquet(ctx, *obj)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write parquet")
		return "", err
	}
	key := r.generateTestCaseKey(obj.TestID)
	time := fmt.Sprintf("%d", time.Now().UTC().Unix())
	metadata := map[string]string{
		"author":  userId,
		"created": time,
		"updated": time,
		"name":    "",
	}

	err = r.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload parquet file")
		return "", err
	}

	span.AddEvent("object successfully written and uploaded", trace.WithAttributes(
		attribute.String("key", key),
		attribute.String("type", "testcase"),
	))
	span.SetStatus(codes.Ok, "")

	return key, nil
}

// Read downloads the Parquet file from S3 using the given key and returns the first TestCase found.
// Returns the TestCase or an error.
// nolint:dupl
func (r *testcaseStorageRepository) Read(ctx context.Context, key string) (*entity.TestCase, error) {
	ctx, span := r.tracer.Start(ctx, "testCaseStorageRepository.Read")
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
	if err := validateTestCaseData(&items[0]); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	span.SetStatus(codes.Ok, "")
	return &items[0], nil
}

func (r *testcaseStorageRepository) ReadAllMetadata(ctx context.Context) ([]*entity.TestcaseMetadata, error) {
	ctx, span := r.tracer.Start(ctx, "testCaseStorageRepository.ReadAllMetadata")
	defer span.End()

	fileKeys, err := r.s3Wrapper.ListParquetFiles(ctx, r.s3Prefix)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to list all parquet files")
		return nil, fmt.Errorf("failed to list all parquet files: %w", err)
	}

	allMetadata := make([]*entity.TestcaseMetadata, 0, len(fileKeys))

	for _, fileKey := range fileKeys {
		_, metadata, err := r.s3Wrapper.DownloadParquetFile(ctx, fileKey)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("failed to download parquet file with key: %s", fileKey))
			span.AddEvent("download failed", trace.WithAttributes(
				attribute.String("fileKey", fileKey),
			))
			r.logger.Warn("failed to download parquet file", "fileKey", fileKey, "error", err)
			continue
		}

		span.AddEvent("parquet file downloaded", trace.WithAttributes(
			attribute.String("fileKey", fileKey),
		))

		allMetadata = append(allMetadata, &entity.TestcaseMetadata{
			Key:     fileKey,
			Author:  metadata["author"],
			Created: metadata["created"],
			Updated: metadata["updated"],
			Name:    metadata["name"],
		})
	}

	span.SetStatus(codes.Ok, "")
	return allMetadata, nil
}

// Update overwrites the existing Parquet file at the given key with the provided TestCase.
// Returns an error if the key does not exist or the operation fails.
// nolint:dupl
func (r *testcaseStorageRepository) Update(ctx context.Context, obj *entity.TestCase, key string) error {
	ctx, span := r.tracer.Start(ctx, "testCaseStorageRepository.Update")
	defer span.End()

	if err := assert.StringNotEmpty(key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return fmt.Errorf("key must not be empty")
	}

	if err := validateTestCaseData(obj); err != nil {
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
		err := errors.New("cannot update: key does not exist")
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
		attribute.String("type", "testcase"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
// Returns an error if the deletion fails.
// nolint:dupl
func (r *testcaseStorageRepository) Delete(ctx context.Context, key string) error {
	ctx, span := r.tracer.Start(ctx, "testCaseStorageRepository.Delete")
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
		attribute.String("type", "testcase"),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// generateTestCaseKey creates a unique S3 key for a TestCase object.
// The format is: "testcase/<testCaseID>_<timestamp>.parquet"
func (r *testcaseStorageRepository) generateTestCaseKey(testcaseId string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s%s_%d.parquet", r.s3Prefix, testcaseId, timestamp)
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
	if err := assert.StringNotEmpty(testcase.TestCode.Code); err != nil {
		return fmt.Errorf("testcase.TestCode.Code must not be empty: %w", err)
	}
	return nil
}
