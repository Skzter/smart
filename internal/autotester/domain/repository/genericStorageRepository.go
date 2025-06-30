package repository

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// GenericStorageRepository provides a generic implementation of StorageRepository
// for any entity type T, using S3 and Parquet wrappers.
type GenericStorageRepository[T any] struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[T]
	logger         *slog.Logger
}

// NewHistoryStorageRepository creates a StorageRepository specifically
// for SessionSummary entities. It uses a generic implementation internally.
func NewHistoryStorageRepository(logger *slog.Logger) (StorageRepository[entity.SessionSummary], error) {
	repo, err := newGenericStorageRepository[entity.SessionSummary](logger)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// NewTestCaseStorageRepository creates a StorageRepository specifically
// for TestCase entities. It uses a generic implementation internally.
func NewTestCaseStorageRepository(logger *slog.Logger) (StorageRepository[entity.TestCase], error) {
	repo, err := newGenericStorageRepository[entity.TestCase](logger)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// newGenericStorageRepository constructs a GenericStorageRepository instance for type T.
// It is unexported to enforce usage of specific typed constructors.
func newGenericStorageRepository[T any](logger *slog.Logger) (*GenericStorageRepository[T], error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	s3Config := wrapperEntity.S3Config{
		Bucket:    "autotester",
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}

	s3Wrapper, err := service.NewS3Wrapper(logger, s3Config)
	if err != nil {
		return nil, err
	}

	parquetWrapper, err := service.NewParquetWrapper[T](logger, service.DefaultParquetConfig())
	if err != nil {
		return nil, err
	}

	return &GenericStorageRepository[T]{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// Create serializes the given object to Parquet format and uploads it to S3,
// returning the generated key or an error.
func (gsr *GenericStorageRepository[T]) Create(ctx context.Context, obj *T) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("obj must not be nil")
	}

	parquetData, err := gsr.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		gsr.logger.Error("create: writing struct to parquet failed",
			slog.String("type", reflect.TypeOf(obj).Elem().Name()),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	key := generateKey(obj)
	metadata := map[string](string){}

	err = gsr.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		gsr.logger.Error("create: uploading parquet file to S3 failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	gsr.logger.Info("create: object successfully written and uploaded",
		slog.String("key", key),
		slog.String("type", reflect.TypeOf(obj).Elem().Name()),
	)

	return key, nil
}

// Read downloads the Parquet file from S3 using the given key,
// deserializes it, and returns the first object found.
func (gsr *GenericStorageRepository[T]) Read(ctx context.Context, key string) (*T, error) {
	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}

	data, _, err := gsr.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		gsr.logger.Error("read: downloading parquet failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	items, err := gsr.parquetWrapper.ReadStructsFromParquet(data)
	if err != nil {
		gsr.logger.Error("read: parsing parquet data failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no data found for key %s", key)
	}
	return &items[0], nil
}

// Update overwrites the existing Parquet file at the given key with the
// serialized form of the provided object. Returns an error if key does not exist.
func (gsr *GenericStorageRepository[T]) Update(ctx context.Context, obj *T, key string) error {
	if obj == nil {
		return fmt.Errorf("obj must not be nil")
	}
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}

	exists, err := gsr.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		gsr.logger.Error("update: failed to check existence",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to check if key exists: %w", err)
	}
	if !exists {
		gsr.logger.Error("update: key does not exist, aborting",
			slog.String("key", key),
		)
		return fmt.Errorf("cannot update: key does not exist")
	}

	parquetData, err := gsr.parquetWrapper.WriteStructToParquet(*obj)
	if err != nil {
		gsr.logger.Error("update: writing struct to parquet failed",
			slog.String("type", reflect.TypeOf(obj).Elem().Name()),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to serialize object: %w", err)
	}

	metadata := map[string]string{}

	err = gsr.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		return fmt.Errorf("failed to upload updated object: %w", err)
	}

	gsr.logger.Info("update: object successfully overwritten",
		slog.String("key", key),
		slog.String("type", reflect.TypeOf(obj).Elem().Name()),
	)

	return nil
}

// Delete removes the Parquet file associated with the given key from S3.
func (gsr *GenericStorageRepository[T]) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key must not be empty")
	}

	if err := gsr.s3Wrapper.DeleteParquetFile(ctx, key); err != nil {
		gsr.logger.Error("delete: removing parquet failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return err
	}

	gsr.logger.Info("delete: object successfully deleted",
		slog.String("key", key),
	)
	return nil
}

// generateKey creates an S3 key using the type name of the object as folder
// and a filename with type name and current timestamp.
func generateKey(obj any) string {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	timestamp := time.Now().Format("20060102150405000")
	return fmt.Sprintf("%s/%s_%s", objType.Name(), objType.Name(), timestamp)
}
