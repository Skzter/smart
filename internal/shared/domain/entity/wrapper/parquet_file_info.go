package entity

import "github.com/parquet-go/parquet-go"

// ParquetFileInfo contains metadata about a parquet file
type ParquetFileInfo struct {
	NumRows      int64           `json:"num_rows"`
	NumColumns   int64           `json:"num_columns"`
	NumRowGroups int64           `json:"num_row_groups"`
	FileSize     int64           `json:"file_size_bytes"`
	Compression  string          `json:"compression"`
	CreatedBy    string          `json:"created_by,omitempty"`
	Schema       *parquet.Schema `json:"-"` // Schema is not JSON serializable
}
