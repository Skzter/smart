package repository

import (
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

type DatabaseRepository interface {
	// creates a request into the database
	CreateRequest(entity.DatabaseEntry) error
	// gets Request by ?
	ReadRequest(string) (entity.DatabaseEntry, error)
	// changes the content of a request
	UpdateRequest(string, entity.DatabaseEntry) error
	// deletes a request from the database
	DeleteRequest(string) error
}

type databaseRepository struct {
	s3Wrapper      *service.S3Wrapper
	parquetWrapper *service.ParquetWrapper[entity.DatabaseEntry]
	logger         *slog.Logger
}

// NewDatabaseRepository creates a new DatabaseRepository
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
func (dbR *databaseRepository) CreateRequest(dbEntry entity.DatabaseEntry) error {
	if err := validateDbEntry(dbEntry, dbR); err != nil {
		dbR.logger.Error("Failed to validate dbEntry", slog.String("error", err.Error()))
	}

	parquetData, err := dbR.parquetWrapper.WriteStructToParquet(dbEntry)
	if err != nil {
		dbR.logger.Error("Failed to write parquet", slog.String("error", err.Error()))
		return err
	}

	dbR.logger.Info("Parquet data created", slog.Int("size_bytes", len(parquetData)))

	return nil
}

func (dbR *databaseRepository) ReadRequest(string) (entity.DatabaseEntry, error) {
	return entity.DatabaseEntry{}, nil
}

func (dbR *databaseRepository) UpdateRequest(string, entity.DatabaseEntry) error {
	return nil
}

func (dbR *databaseRepository) DeleteRequest(string) error {
	return nil
}

// validateDbEntry validates dbEntry
func validateDbEntry(dbEntry entity.DatabaseEntry, dbR *databaseRepository) error {
	if err := validateRequest(dbEntry.Request); err != nil {
		dbR.logger.Error("Failed to validate request", slog.String("error", err.Error()))
		return err
	}
	if err := validateResponse(dbEntry.Response); err != nil {
		dbR.logger.Error("Failed to validate response", slog.String("error", err.Error()))
		return err
	}
	if err := validateTags(dbEntry.Tags); err != nil {
		dbR.logger.Error("Failed to validate tags", slog.String("error", err.Error()))
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
