package repository

import (
	"context"
	"fmt"
	"log/slog"

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
	return "", nil
}

func (gsr *GenericStorageRepository[T]) Read(key string) (*T, error) {
	return nil, nil
}

func (gsr *GenericStorageRepository[T]) Update(ctx context.Context, obj *T, key string) (string, error) {
	return "", nil
}

func (gsr *GenericStorageRepository[T]) Delete(key string) error {
	return nil
}
