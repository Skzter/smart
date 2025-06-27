package repository

import (
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// s3StorageRepository implements StorageRepository for AWS S3 storage.
type s3StorageRepository struct {
	s3Wrapper *service.S3Wrapper
	logger    *slog.Logger
}

func NewS3StorageRepository(logger *slog.Logger) (StorageRepository, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, fmt.Errorf("logger cannot be nil: %w", err)
	}

	s3Wrapper, err := service.NewS3Wrapper(logger, entity.S3Config{
		Bucket:    "bucket name",
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3Wrapper: %w", err)
	}

	return &s3StorageRepository{
		s3Wrapper: s3Wrapper,
		logger:    logger,
	}, nil
}

func (s3 *s3StorageRepository) Create() {
	// UploadParquetFile(ctx, key, data, metadata)

	// Fragen für Johannes:
	// soll key hereingegeben werden oder hier erzeugt werden
	// wenn hereingegeben validieren das es den noch nicht gab
}

func (s3 *s3StorageRepository) Read(key string) {
	// DownloadParquetFile(ctx, key)

	// key erst validieren mit Exists, dann lesen
}

func (s3 *s3StorageRepository) Update(key string) error {
	// UploadParquetFile(ctx, key, data, metadata)
	// with existing key, overwrites the old file

	// key erst validieren mit Exists, dann überschreiben
	return nil
}

func (s3 *s3StorageRepository) Delete(key string) error {
	// DeleteParquetFile(ctx, key)

	// key erst validieren mit Exists, dann löschen
	return nil
}
