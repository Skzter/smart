package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/gzip"
	"github.com/parquet-go/parquet-go/compress/snappy"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

var (
	// ErrEmptyParquetData indicates that the provided Parquet data is empty
	// when data was expected.
	ErrEmptyParquetData = errors.New("parquet data cannot be empty")

	// ErrFailedToWriteData indicates that writing data to a Parquet file or buffer failed.
	ErrFailedToWriteData = errors.New("failed to write data to parquet")

	// ErrFailedToCloseWriter indicates that closing the Parquet writer failed.
	ErrFailedToCloseWriter = errors.New("failed to close parquet writer")

	// ErrFailedToReadData indicates that reading data from a Parquet source failed.
	ErrFailedToReadData = errors.New("failed to read parquet data")

	// ErrFailedToCloseReader indicates that closing the Parquet reader failed.
	ErrFailedToCloseReader = errors.New("failed to close parquet reader")

	// ErrInvalidStructType indicates that the provided data is not a struct type.
	// The error message includes the actual type.
	ErrInvalidStructType = errors.New("data must be a struct, got %s")
)

// ParquetFileWrapper defines an interface for working with Parquet serialization and deserialization for a generic struct type T.
type ParquetFileWrapper[T any] interface {
	// WriteStructsToParquet serializes a slice of structs into a Parquet-encoded byte slice.
	WriteStructsToParquet(ctx context.Context, data []T) ([]byte, error)

	// ReadStructsFromParquet deserializes a Parquet-encoded byte slice into a slice of structs.
	ReadStructsFromParquet(ctx context.Context, parquetData []byte) ([]T, error)

	// WriteStructToParquet serializes a single struct into a Parquet-encoded byte slice.
	WriteStructToParquet(ctx context.Context, data T) ([]byte, error)

	// ValidateStruct checks whether the struct conforms to the expected schema for Parquet serialization.
	ValidateStruct(ctx context.Context, data T) error

	// GetParquetFileInfo extracts metadata information (e.g., row count, schema) from a Parquet-encoded byte slice.
	GetParquetFileInfo(ctx context.Context, parquetData []byte) (*entity.ParquetFileInfo, error)

	// GetParquetSchema returns the Parquet schema for the struct type T.
	GetParquetSchema(ctx context.Context) (*parquet.Schema, error)
}

// ParquetWrapper provides generic methods to handle parquet files with Go structs
type ParquetWrapper[T any] struct {
	logger *slog.Logger
	config entity.ParquetConfig
	tracer trace.Tracer
}

// DefaultParquetConfig returns a default configuration for parquet operations
func DefaultParquetConfig() entity.ParquetConfig {
	return entity.ParquetConfig{
		CompressionCodec: "snappy",
		PageSize:         8 * 1024, // 8KB
		RowGroupSize:     1000,     // 1000 rows per group
	}
}

// NewParquetWrapper creates a new ParquetWrapper instance
func NewParquetWrapper[T any](logger *slog.Logger, config entity.ParquetConfig, tracer trace.Tracer) (ParquetFileWrapper[T], error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, ErrNilLogger
	}

	wrapper := &ParquetWrapper[T]{
		logger: logger,
		config: config,
		tracer: tracer,
	}

	logger.Debug("ParquetWrapper initialized")

	return wrapper, nil
}

// getCompressionOptions returns parquet writer options based on compression codec
func getCompressionOptions(codec string) []parquet.WriterOption {
	var options []parquet.WriterOption

	switch codec {
	case "snappy":
		options = append(options, parquet.Compression(&snappy.Codec{}))
	case "gzip":
		options = append(options, parquet.Compression(&gzip.Codec{}))
	case "none", "":
		// No compression
	default:
		// Default to snappy
		options = append(options, parquet.Compression(&snappy.Codec{}))
	}

	return options
}

// WriteStructsToParquet converts a slice of structs to parquet format
func (p *ParquetWrapper[T]) WriteStructsToParquet(ctx context.Context, data []T) ([]byte, error) {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.WriteStructsToParquet")
	defer span.End()

	if len(data) == 0 {
		err := ErrEmptyData
		span.RecordError(err)
		span.SetStatus(codes.Error, "data validation failed")
		p.logger.Error("Data validation failed",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrValidation
	}

	// Get the type of the struct
	structType := reflect.TypeOf(data[0])
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	p.logger.Debug("Writing structs to parquet",
		slog.String("type", structType.Name()),
		slog.Int("count", len(data)),
		slog.String("compression", p.config.CompressionCodec),
	)

	// Create a buffer to write parquet data
	var buf bytes.Buffer

	// Get compression options
	options := getCompressionOptions(p.config.CompressionCodec)

	// Create parquet file writer with configuration
	writer := parquet.NewGenericWriter[T](&buf, options...)

	// Write the data
	_, err := writer.Write(data)
	if err != nil {
		err := ErrFailedToWriteData
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write data to parquet")
		p.logger.Error("Failed to write data to parquet",
			slog.String("type", structType.Name()),
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrInternalServer
	}

	// Close the writer to finalize the file
	err = writer.Close()
	if err != nil {
		err := ErrFailedToCloseWriter
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close parquet writer")
		p.logger.Error("Failed to close parquet writer",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrInternalServer
	}

	parquetData := buf.Bytes()

	p.logger.Debug("Successfully wrote structs to parquet",
		slog.String("type", structType.Name()),
		slog.Int("count", len(data)),
		slog.Int("size_bytes", len(parquetData)),
	)
	span.SetStatus(codes.Ok, "")

	return parquetData, nil
}

// ReadStructsFromParquet reads parquet data and converts it to a slice of structs
func (p *ParquetWrapper[T]) ReadStructsFromParquet(ctx context.Context, parquetData []byte) ([]T, error) {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.ReadStructsFromParquet")
	defer span.End()

	if len(parquetData) == 0 {
		err := ErrEmptyParquetData
		span.RecordError(err)
		span.SetStatus(codes.Error, "data validation failed")
		p.logger.Error("Data validation failed",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrValidation
	}

	// Get the type of the struct for logging
	var zero T
	structType := reflect.TypeOf(zero)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	p.logger.Debug("Reading structs from parquet",
		slog.String("type", structType.Name()),
		slog.Int("data_size_bytes", len(parquetData)),
	)

	// Create a reader from the parquet data
	reader := bytes.NewReader(parquetData)
	parquetReader := parquet.NewGenericReader[T](reader)

	// Read all rows
	result := make([]T, 0)

	// Read all data at once
	for {
		rows := make([]T, 100) // Read in smaller batches
		n, err := parquetReader.Read(rows)
		if err != nil && err != io.EOF {
			err := ErrFailedToReadData
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to read parquet data")
			p.logger.Error("Failed to read parquet data",
				slog.String("type", structType.Name()),
				slog.String("error", err.Error()),
			)
			return nil, sharedErrors.ErrInternalServer
		}

		if n > 0 {
			result = append(result, rows[:n]...)
		}

		if err == io.EOF || n == 0 {
			break
		}
	}

	// Close the reader
	err := parquetReader.Close()
	if err != nil {
		err := ErrFailedToCloseReader
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close parquet reader")
		p.logger.Error("Failed to close parquet reader",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrInternalServer
	}

	span.AddEvent("Structs read from parquet", trace.WithAttributes(
		attribute.String("type", structType.Name()),
		attribute.Int("count", len(result)),
	))
	span.SetStatus(codes.Ok, "")

	p.logger.Debug("Successfully read structs from parquet",
		slog.String("type", structType.Name()),
		slog.Int("count", len(result)),
	)

	return result, nil
}

// WriteStructToParquet converts a single struct to parquet format (convenience function)
func (p *ParquetWrapper[T]) WriteStructToParquet(ctx context.Context, data T) ([]byte, error) {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.WriteStructToParquet")
	defer span.End()

	var zero T
	structType := reflect.TypeOf(zero)

	result, err := p.WriteStructsToParquet(ctx, []T{data})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to write struct to parquet")
		p.logger.Error("Failed to write struct to parquet",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrInternalServer
	}

	span.AddEvent("Struct written to parquet", trace.WithAttributes(
		attribute.String("type", structType.Name()),
	))

	span.SetStatus(codes.Ok, "")
	return result, nil
}

// GetParquetSchema returns the parquet schema for a given struct type
func (p *ParquetWrapper[T]) GetParquetSchema(ctx context.Context) (*parquet.Schema, error) {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.GetParquetSchema")
	defer span.End()

	var zero T
	structType := reflect.TypeOf(zero)

	p.logger.Debug("Getting parquet schema",
		slog.String("type", structType.Name()),
	)

	// Create a temporary buffer and writer to get the schema
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[T](&buf)
	schema := writer.Schema()
	err := writer.Close()
	if err != nil {
		err := ErrFailedToCloseWriter
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to close parquet writer")
		p.logger.Error("Failed to close parquet writer",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrInternalServer
	}

	span.AddEvent("Parquet schema generated", trace.WithAttributes(
		attribute.String("type", structType.Name()),
		attribute.Int("fields", len(schema.Fields())),
	))
	span.SetStatus(codes.Ok, "")

	p.logger.Debug("Generated parquet schema",
		slog.String("type", structType.Name()),
		slog.Int("fields", len(schema.Fields())),
	)

	return schema, nil
}

// ValidateStruct validates that a struct is suitable for parquet serialization
func (p *ParquetWrapper[T]) ValidateStruct(ctx context.Context, data T) error {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.ValidateStruct")
	defer span.End()

	structType := reflect.TypeOf(data)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		err := fmt.Errorf(ErrInvalidStructType.Error(), structType.Kind())
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid struct type")
		p.logger.Error(err.Error())
		return sharedErrors.ErrValidation
	}

	p.logger.Debug("Validating struct for parquet compatibility",
		slog.String("type", structType.Name()),
		slog.Int("fields", structType.NumField()),
	)

	// Check if struct has exported fields
	hasExportedFields := false
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.IsExported() {
			hasExportedFields = true
			break
		}
	}

	if !hasExportedFields {
		err := fmt.Errorf(ErrInvalidStructType.Error(), structType.Kind())
		span.RecordError(err)
		span.SetStatus(codes.Error, "struct has no exported fields")
		p.logger.Error(err.Error())
		return sharedErrors.ErrValidation
	}

	span.AddEvent("Struct validated", trace.WithAttributes(
		attribute.String("type", structType.Name()),
	))

	span.SetStatus(codes.Ok, "")
	return nil
}

// GetParquetFileInfo extracts information from parquet data without fully reading it
func (p *ParquetWrapper[T]) GetParquetFileInfo(ctx context.Context, parquetData []byte) (*entity.ParquetFileInfo, error) {
	_, span := p.tracer.Start(ctx, "ParquetWrapper.GetParquetFileInfo")
	defer span.End()

	if len(parquetData) == 0 {
		err := ErrEmptyParquetData
		span.RecordError(err)
		span.SetStatus(codes.Error, "data validation failed")
		p.logger.Error("Data validation failed",
			slog.String("error", err.Error()),
		)
		return nil, sharedErrors.ErrValidation
	}

	reader := bytes.NewReader(parquetData)

	// Create a generic reader to access file metadata
	file, err := parquet.OpenFile(reader, int64(len(parquetData)))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to open parquet file")
		return nil, fmt.Errorf("failed to open parquet file: %w", err)
	}

	schema := file.Schema()
	metadata := file.Metadata()

	info := &entity.ParquetFileInfo{
		NumRows:      metadata.NumRows,
		NumColumns:   int64(len(schema.Fields())),
		NumRowGroups: int64(len(metadata.RowGroups)),
		FileSize:     int64(len(parquetData)),
		Schema:       schema,
		CreatedBy:    metadata.CreatedBy,
	}

	// Extract compression information from first row group if available
	if len(metadata.RowGroups) > 0 && len(metadata.RowGroups[0].Columns) > 0 {
		// Simplified compression detection
		info.Compression = "unknown"
	}

	p.logger.Debug("Extracted parquet file info",
		slog.Int64("rows", info.NumRows),
		slog.Int64("columns", info.NumColumns),
		slog.Int64("row_groups", info.NumRowGroups),
		slog.Int64("size_bytes", info.FileSize),
		slog.String("compression", info.Compression),
	)

	var zero T
	structType := reflect.TypeOf(zero)

	span.AddEvent("Parquet file info extracted", trace.WithAttributes(
		attribute.String("type", structType.Name()),
		attribute.Int64("rows", info.NumRows),
		attribute.Int64("columns", info.NumColumns),
		attribute.Int64("row_groups", info.NumRowGroups),
		attribute.Int64("size_bytes", info.FileSize),
		attribute.String("compression", info.Compression),
	))

	span.SetStatus(codes.Ok, "")
	return info, nil
}
