package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	domainRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository"
)

// DatabaseRepository struct defines the infrastructure layer of database
type databaseRepository struct {
	s3      service.S3StorageWrapper
	parquet service.ParquetFileWrapper[entity.DatabaseEntry]
	log     *slog.Logger
	tracer  trace.Tracer
	prefix  string
}

// NewDatabaseRepository creates a new instance of DB repo
func NewDatabaseRepository(
	log *slog.Logger,
	s3 service.S3StorageWrapper,
	parquet service.ParquetFileWrapper[entity.DatabaseEntry],
	tracer trace.Tracer,
	prefix string,
) domainRepo.DatabaseRepository {
	return &databaseRepository{
		s3:      s3,
		parquet: parquet,
		log:     log,
		tracer:  tracer,
		prefix:  prefix,
	}
}

// CreateRequest writes the given entry to the database by converting it to Parquet format and uploading it to S3.
func (r *databaseRepository) CreateRequest(ctx context.Context, entry entity.DatabaseEntry) error {
	if err := domainRepo.ValidateDatabaseEntry(entry); err != nil {
		return err
	}

	data, err := r.parquet.WriteStructToParquet(ctx, entry)
	if err != nil {
		return err
	}

	ts := fmt.Sprintf("%d", time.Now().Unix())
	key := domainRepo.GenerateKey(entry.Tags, ts)

	return r.s3.UploadParquetFile(ctx, r.prefix+key, data, map[string]string{
		"created": ts,
	})
}

// ReadRequest retrieves a request from the database by its key, downloading the Parquet file and reading its content.
func (r *databaseRepository) ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error) {
	data, _, err := r.s3.DownloadParquetFile(ctx, r.prefix+key)
	if err != nil {
		return nil, err
	}

	entries, err := r.parquet.ReadStructsFromParquet(ctx, data)
	if err != nil {
		return nil, err
	}
	return &entries[0], nil
}

// UpdateRequest updates an existing request in the database by downloading the Parquet file, modifying its content, and re-uploading it.
func (r *databaseRepository) UpdateRequest(ctx context.Context, key string, entry entity.DatabaseEntry) error {
	data, err := r.parquet.WriteStructToParquet(ctx, entry)
	if err != nil {
		return err
	}
	return r.s3.UploadParquetFile(ctx, r.prefix+key, data, map[string]string{
		"updated": fmt.Sprintf("%d", time.Now().Unix()),
	})
}

// DeleteRequest deletes a request from the database by removing the Parquet file associated with the given key.
func (r *databaseRepository) DeleteRequest(ctx context.Context, key string) error {
	return r.s3.DeleteParquetFile(ctx, r.prefix+key)
}

func (r *databaseRepository) ListAllKeys(ctx context.Context) ([]string, error) {
	return r.s3.ListParquetFiles(ctx, r.prefix)
}
