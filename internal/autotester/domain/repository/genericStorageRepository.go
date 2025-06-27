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
	s3Wrapper      *service.S3Wrapper
	parquetWrapper *service.ParquetWrapper[T]
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
		Bucket:    "bucket name",
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

func (gsr *GenericStorageRepository[T]) Read(key string) (*T, error) {
	return nil, nil
}

func (gsr *GenericStorageRepository[T]) Update(ctx context.Context, obj *T, key string) (string, error) {
	if obj == nil {
		return "", fmt.Errorf("obj must not be nil")
	}

	err := gsr.Delete(key)
	if err != nil {
		gsr.logger.Error("update: deleting existing object failed",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("failed to delete existing object: %w", err)
	}

	newKey, err := gsr.Create(ctx, obj)
	if err != nil {
		gsr.logger.Error("update: creating updated object failed",
			slog.String("original_key", key),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("failed to create updated object: %w", err)
	}

	gsr.logger.Info("update: object successfully deleted and replaced",
		slog.String("old_key", key),
		slog.String("new_key", newKey),
		slog.String("type", reflect.TypeOf(obj).Elem().Name()),
	)

	return newKey, nil
}

func (gsr *GenericStorageRepository[T]) Delete(key string) error {
	return nil
}

// currently only generates the filename part of the path/filename
func generateKey(obj any) string {
	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}
	time := time.Now().Format("20060102150405000")
	return fmt.Sprintf("%s_%s", objType.Name(), time)
}
