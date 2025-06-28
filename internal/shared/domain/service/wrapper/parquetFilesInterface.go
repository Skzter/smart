package service

import (
	"github.com/parquet-go/parquet-go"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
)

// ParquetFileWrapper defines methods for serializing and deserializing structs to and from Parquet format.
type ParquetFileWrapper[T any] interface {
	// WriteStructsToParquet serializes a slice of structs into a Parquet-encoded byte slice.
	WriteStructsToParquet(data []T) ([]byte, error)

	// ReadStructsFromParquet deserializes a Parquet-encoded byte slice into a slice of structs.
	ReadStructsFromParquet(parquetData []byte) ([]T, error)

	// WriteStructToParquet serializes a single struct into a Parquet-encoded byte slice.
	WriteStructToParquet(data T) ([]byte, error)

	// ValidateStruct checks whether the struct conforms to the expected schema for Parquet serialization.
	ValidateStruct(data T) error

	// GetParquetFileInfo extracts metadata information (e.g., row count, schema) from a Parquet-encoded byte slice.
	GetParquetFileInfo(parquetData []byte) (*entity.ParquetFileInfo, error)

	// GetParquetSchema returns the Parquet schema for the struct type T.
	GetParquetSchema() (*parquet.Schema, error)
}
