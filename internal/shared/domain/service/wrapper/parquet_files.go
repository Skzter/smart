package service

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"reflect"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/gzip"
	"github.com/parquet-go/parquet-go/compress/snappy"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ParquetWrapper provides generic methods to handle parquet files with Go structs
type ParquetWrapper[T any] struct {
	logger *slog.Logger
	config entity.ParquetConfig
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
func NewParquetWrapper[T any](logger *slog.Logger, config entity.ParquetConfig) (*ParquetWrapper[T], error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, fmt.Errorf("logger cannot be nil: %w", err)
	}

	wrapper := &ParquetWrapper[T]{
		logger: logger,
		config: config,
	}

	logger.Info("ParquetWrapper initialized")

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
func (p *ParquetWrapper[T]) WriteStructsToParquet(data []T) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data slice cannot be empty")
	}

	// Get the type of the struct
	structType := reflect.TypeOf(data[0])
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	p.logger.Info("Writing structs to parquet",
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
		p.logger.Error("Failed to write data to parquet",
			slog.String("type", structType.Name()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to write data to parquet: %w", err)
	}

	// Close the writer to finalize the file
	err = writer.Close()
	if err != nil {
		p.logger.Error("Failed to close parquet writer",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	parquetData := buf.Bytes()

	p.logger.Info("Successfully wrote structs to parquet",
		slog.String("type", structType.Name()),
		slog.Int("count", len(data)),
		slog.Int("size_bytes", len(parquetData)),
	)

	return parquetData, nil
}

// ReadStructsFromParquet reads parquet data and converts it to a slice of structs
func (p *ParquetWrapper[T]) ReadStructsFromParquet(parquetData []byte) ([]T, error) {
	if len(parquetData) == 0 {
		return nil, fmt.Errorf("parquet data cannot be empty")
	}

	// Get the type of the struct for logging
	var zero T
	structType := reflect.TypeOf(zero)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	p.logger.Info("Reading structs from parquet",
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
			p.logger.Error("Failed to read parquet data",
				slog.String("type", structType.Name()),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to read parquet data: %w", err)
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
		p.logger.Error("Failed to close parquet reader",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to close parquet reader: %w", err)
	}

	p.logger.Info("Successfully read structs from parquet",
		slog.String("type", structType.Name()),
		slog.Int("count", len(result)),
	)

	return result, nil
}

// WriteStructToParquet converts a single struct to parquet format (convenience function)
func (p *ParquetWrapper[T]) WriteStructToParquet(data T) ([]byte, error) {
	return p.WriteStructsToParquet([]T{data})
}

// GetParquetSchema returns the parquet schema for a given struct type
func (p *ParquetWrapper[T]) GetParquetSchema() (*parquet.Schema, error) {
	var zero T
	structType := reflect.TypeOf(zero)

	p.logger.Info("Getting parquet schema",
		slog.String("type", structType.Name()),
	)

	// Create a temporary buffer and writer to get the schema
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[T](&buf)
	schema := writer.Schema()
	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	p.logger.Info("Generated parquet schema",
		slog.String("type", structType.Name()),
		slog.Int("fields", len(schema.Fields())),
	)

	return schema, nil
}

// ValidateStruct validates that a struct is suitable for parquet serialization
func (p *ParquetWrapper[T]) ValidateStruct(data T) error {
	structType := reflect.TypeOf(data)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct, got %s", structType.Kind())
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
		return fmt.Errorf("struct must have at least one exported field")
	}

	return nil
}

// GetParquetFileInfo extracts information from parquet data without fully reading it
func (p *ParquetWrapper[T]) GetParquetFileInfo(parquetData []byte) (*entity.ParquetFileInfo, error) {
	if len(parquetData) == 0 {
		return nil, fmt.Errorf("parquet data cannot be empty")
	}

	reader := bytes.NewReader(parquetData)

	// Create a generic reader to access file metadata
	file, err := parquet.OpenFile(reader, int64(len(parquetData)))
	if err != nil {
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

	p.logger.Info("Extracted parquet file info",
		slog.Int64("rows", info.NumRows),
		slog.Int64("columns", info.NumColumns),
		slog.Int64("row_groups", info.NumRowGroups),
		slog.Int64("size_bytes", info.FileSize),
		slog.String("compression", info.Compression),
	)

	return info, nil
}
