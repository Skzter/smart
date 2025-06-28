package service

import "context"

// S3StorageWrapper defines the operations for interacting with Parquet files in S3-compatible storage.
type S3StorageWrapper interface {
	// UploadParquetFile uploads a Parquet file to the given key with optional metadata.
	UploadParquetFile(ctx context.Context, key string, data []byte, metadata map[string]string) error

	// DownloadParquetFile retrieves a Parquet file and its metadata from the given key.
	DownloadParquetFile(ctx context.Context, key string) ([]byte, map[string]string, error)

	// ListParquetFiles lists all file keys with the given prefix in the S3 bucket.
	ListParquetFiles(ctx context.Context, prefix string) ([]string, error)

	// DeleteParquetFile removes a file from the S3 bucket by its key.
	DeleteParquetFile(ctx context.Context, key string) error

	// FileExists checks whether a file with the given key exists in the S3 bucket.
	FileExists(ctx context.Context, key string) (bool, error)

	// GetFileSize returns the size of the file in bytes for the given key.
	GetFileSize(ctx context.Context, key string) (int64, error)
}
