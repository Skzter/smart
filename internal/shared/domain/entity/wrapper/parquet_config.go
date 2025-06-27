package entity

// ParquetConfig holds configuration for parquet file operations
type ParquetConfig struct {
	CompressionCodec string // Compression codec ("snappy", "gzip", "none")
	PageSize         int64  // Page size in bytes (default: 8KB)
	RowGroupSize     int64  // Row group size in rows (default: 1000)
}
