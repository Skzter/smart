package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

type DatabaseRepository interface {
	// creates a request into the database
	CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error
	// gets Request (access through key)
	ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error)
	// changes the content of a request (access through key)
	UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error
	// deletes a request from the database (access through key)
	DeleteRequest(ctx context.Context, key string) error
}

type databaseRepository struct {
	s3Wrapper      service.S3StorageWrapper
	parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry]
	logger         *slog.Logger
}

// NewDatabaseRepository creates a new instance of DatabaseRepository
func NewDatabaseRepository(logger *slog.Logger) (DatabaseRepository, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	s3Config := wrapperEntity.S3Config{
		Bucket:    "suproxy",
		AccessKey: build.AwsAccessKey,
		SecretKey: build.AwsSecretAccessKey,
	}

	s3Wrapper, err := service.NewS3Wrapper(logger, s3Config)
	if err != nil {
		return nil, err
	}

	parquetWrapper, err := service.NewParquetWrapper[entity.DatabaseEntry](logger, service.DefaultParquetConfig())
	if err != nil {
		return nil, err
	}

	return &databaseRepository{
		s3Wrapper:      s3Wrapper,
		parquetWrapper: parquetWrapper,
		logger:         logger,
	}, nil
}

// CreateRequest creates, converts JSON to parquet and uploads request to database
func (dbR *databaseRepository) CreateRequest(ctx context.Context, dbEntry entity.DatabaseEntry) error {
	if err := validateDbEntry(dbEntry, dbR); err != nil {
		dbR.logger.Error("failed to validate dbEntry", slog.String("error", err.Error()))
		return err
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(dbEntry)
	if err != nil {
		dbR.logger.Error("failed to write parquet", slog.String("error", err.Error()))
		return err
	}

	dbR.logger.Info("parquet data created", slog.Int("size_bytes", len(parquetData)))

	metadata := map[string]string{
		"created": string(rune(time.Now().Unix())),
	}

	key := generateKey(dbEntry.Tags[0], dbEntry.Tags[1], metadata["created"])

	// Upload file (automatically adds .parquet extension if missing)
	err = dbR.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		dbR.logger.Error("upload failed", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// ReadRequest reads a request from the database (access through key)
func (dbR *databaseRepository) ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error) {
	parquetData, metadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		dbR.logger.Error("download failed", slog.String("error", err.Error()))
		return nil, err
	}
	dbR.logger.Info("file downloaded",
		slog.Int("size", len(parquetData)),
		slog.Any("metadata", metadata),
	)
	dbEntries, err := dbR.parquetWrapper.ReadStructsFromParquet(parquetData)
	if err != nil {
		dbR.logger.Error("failed to read parquet data", slog.String("error", err.Error()))
		return nil, err
	}
	dbR.logger.Info("events read from parquet", slog.Int("count", len(dbEntries)))
	firstEntry := dbEntries[0]
	return &firstEntry, nil
}

func (dbR *databaseRepository) UpdateRequest(ctx context.Context, key string, dbEntry entity.DatabaseEntry) error {
	if err := validateDbEntry(dbEntry, dbR); err != nil {
		dbR.logger.Error("failed to validate dbEntry", slog.String("error", err.Error()))
		return err
	}

	_, oldmetadata, err := dbR.s3Wrapper.DownloadParquetFile(ctx, key)
	if err != nil {
		dbR.logger.Error("download failed", slog.String("error", err.Error()))
		return err
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(dbEntry)
	if err != nil {
		dbR.logger.Error("failed to write parquet", slog.String("error", err.Error()))
		return err
	}

	dbR.logger.Info("parquet data created", slog.Int("size_bytes", len(parquetData)))

	metadata := map[string]string{
		"created": string(oldmetadata["created"]),
		"updated": string(rune(time.Now().Unix())),
	}

	// overwrites old dbEntry
	err = dbR.s3Wrapper.UploadParquetFile(ctx, key, parquetData, metadata)
	if err != nil {
		dbR.logger.Error("upload failed", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (dbR *databaseRepository) DeleteRequest(ctx context.Context, key string) error {
	err := dbR.s3Wrapper.DeleteParquetFile(ctx, key)
	if err != nil {
		dbR.logger.Error("failed to delete file", slog.String("error", err.Error()))
		return err
	}

	dbR.logger.Info("file deleted successfully")

	return nil
}

// validateDbEntry validates the dbEntry and logs errors if validation fails
func validateDbEntry(dbEntry entity.DatabaseEntry, dbR *databaseRepository) error {
	if err := validateRequest(dbEntry.Request); err != nil {
		dbR.logger.Error("failed to validate request", slog.String("error", err.Error()))
		return err
	}
	if err := validateResponse(dbEntry.Response); err != nil {
		dbR.logger.Error("failed to validate response", slog.String("error", err.Error()))
		return err
	}
	if err := validateTags(dbEntry.Tags); err != nil {
		dbR.logger.Error("failed to validate tags", slog.String("error", err.Error()))
		return err
	}
	return nil
}

// validateRequest validates request from dbEntry
func validateRequest(rq entity.Request) error {
	if len(rq.Header) == 0 {
		return fmt.Errorf("header must not be empty")
	}

	if err := assert.StringNotEmpty(rq.Prompt); err != nil {
		return fmt.Errorf("prompt must not be empty: %w", err)
	}

	if err := assert.StringNotEmpty(rq.Destination); err != nil {
		return fmt.Errorf("destination must not be empty: %w", err)
	}

	if err := assert.StringNotEmpty(rq.Request); err != nil {
		return fmt.Errorf("request must not be empty: %w", err)
	}
	return nil
}

// validateResponse validates response from dbEntry
func validateResponse(rp entity.Response) error {
	if err := assert.StringNotEmpty(rp.Response); err != nil {
		return fmt.Errorf("response must not be empty: %w", err)
	}
	return nil
}

// validateTags validates tags from dbEntry
func validateTags(t []string) error {
	if len(t) == 0 {
		return fmt.Errorf("tags must not be empty")
	}
	return nil
}

// generateKey generates a unique key for the database entry based on tags and timestamp
func generateKey(tag1 string, tag2 string, unixTimestamp string) string {
	return tag1 + "-" + tag2 + "-" + unixTimestamp
}
