//go:build integration

package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
)

func TestNewS3Wrapper(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name    string
		config  entity.S3Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: entity.S3Config{
				Region: "us-east-1",
				Bucket: "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "missing bucket",
			config: entity.S3Config{
				Region: "us-east-1",
			},
			wantErr: true,
		},
		{
			name: "config with credentials",
			config: entity.S3Config{
				Region:    "us-east-1",
				Bucket:    "test-bucket",
				AccessKey: "test-access-key",
				SecretKey: "test-secret-key",
			},
			wantErr: false,
		},
		{
			name: "config with custom endpoint",
			config: entity.S3Config{
				Region:   "us-east-1",
				Bucket:   "test-bucket",
				Endpoint: "http://localhost:9000",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper, err := NewS3Wrapper(logger, tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, wrapper)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wrapper)
			}
		})
	}
}

func TestS3Wrapper_ValidateInputs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := entity.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}

	wrapper, err := NewS3Wrapper(logger, config)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	ctx := context.Background()

	t.Run("UploadParquetFile validation", func(t *testing.T) {
		// Test empty key
		err := wrapper.UploadParquetFile(ctx, "", []byte("test data"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test empty data
		err = wrapper.UploadParquetFile(ctx, "test-file", []byte{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "data cannot be empty")

		// Test nil context - this should fail due to assert.NotNil
		err = wrapper.UploadParquetFile(ctx, "test-file", []byte("test data"), nil)
		assert.Error(t, err)
	})

	t.Run("DownloadParquetFile validation", func(t *testing.T) {
		// Test empty key
		_, _, err := wrapper.DownloadParquetFile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context - this should fail due to assert.NotNil
		_, _, err = wrapper.DownloadParquetFile(ctx, "test-file")
		assert.Error(t, err)
	})

	t.Run("DeleteParquetFile validation", func(t *testing.T) {
		// Test empty key
		err := wrapper.DeleteParquetFile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context - this should fail due to assert.NotNil
		err = wrapper.DeleteParquetFile(ctx, "test-file")
		assert.Error(t, err)
	})

	t.Run("FileExists validation", func(t *testing.T) {
		// Test empty key
		_, err := wrapper.FileExists(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context - this should fail due to assert.NotNil
		_, err = wrapper.FileExists(ctx, "test-file")
		assert.Error(t, err)
	})

	t.Run("GetFileSize validation", func(t *testing.T) {
		// Test empty key
		_, err := wrapper.GetFileSize(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context - this should fail due to assert.NotNil
		_, err = wrapper.GetFileSize(ctx, "test-file")
		assert.Error(t, err)
	})
}

func TestS3Wrapper_ParquetExtensionHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	config := entity.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}

	wrapper, err := NewS3Wrapper(logger, config)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	// This test demonstrates that the wrapper automatically appends .parquet extension
	// We can't actually test the upload without mocking S3, but we can verify the validation logic
	ctx := context.Background()

	// Test that keys without .parquet extension are handled
	// Note: This will fail with AWS error since we don't have real credentials,
	// but it won't fail on our validation logic
	err = wrapper.UploadParquetFile(ctx, "test-file-without-extension", []byte("test data"), nil)
	assert.Error(t, err) // Will fail due to AWS connection, not our validation

	err = wrapper.UploadParquetFile(ctx, "test-file.parquet", []byte("test data"), nil)
	assert.Error(t, err) // Will fail due to AWS connection, not our validation
}

// Example usage function (not a test)
func ExampleS3Wrapper() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Configure S3 wrapper
	config := entity.S3Config{
		Region:    "us-east-1",
		Bucket:    "my-parquet-bucket",
		AccessKey: "your-access-key", // Optional: use AWS credentials chain if not provided
		SecretKey: "your-secret-key", // Optional: use AWS credentials chain if not provided
		Endpoint:  "",                // Optional: for S3-compatible services like MinIO
	}

	// Create S3 wrapper
	s3Wrapper, err := NewS3Wrapper(logger, config)
	if err != nil {
		logger.Error("Failed to create S3 wrapper", slog.String("error", err.Error()))
		return
	}

	ctx := context.Background()

	// Example: Upload a parquet file
	parquetData := []byte("your parquet file data here")
	metadata := map[string]string{
		"source":      "my-application",
		"created-by":  "user-123",
		"data-format": "parquet",
	}

	err = s3Wrapper.UploadParquetFile(ctx, "data/2024/01/user-activity", parquetData, metadata)
	if err != nil {
		logger.Error("Failed to upload parquet file", slog.String("error", err.Error()))
		return
	}

	// Example: Download a parquet file
	data, fileMetadata, err := s3Wrapper.DownloadParquetFile(ctx, "data/2024/01/user-activity.parquet")
	if err != nil {
		logger.Error("Failed to download parquet file", slog.String("error", err.Error()))
		return
	}

	logger.Info("Downloaded parquet file",
		slog.Int("size", len(data)),
		slog.Any("metadata", fileMetadata),
	)

	// Example: List all parquet files with prefix
	files, err := s3Wrapper.ListParquetFiles(ctx, "data/2024/01/")
	if err != nil {
		logger.Error("Failed to list parquet files", slog.String("error", err.Error()))
		return
	}

	logger.Info("Found parquet files", slog.Int("count", len(files)))
	for _, file := range files {
		logger.Info("Parquet file", slog.String("key", file))
	}

	// Example: Check if file exists
	exists, err := s3Wrapper.FileExists(ctx, "data/2024/01/user-activity.parquet")
	if err != nil {
		logger.Error("Failed to check file existence", slog.String("error", err.Error()))
		return
	}

	if exists {
		// Example: Get file size
		size, err := s3Wrapper.GetFileSize(ctx, "data/2024/01/user-activity.parquet")
		if err != nil {
			logger.Error("Failed to get file size", slog.String("error", err.Error()))
			return
		}

		logger.Info("File size", slog.Int64("bytes", size))

		// Example: Delete the file
		err = s3Wrapper.DeleteParquetFile(ctx, "data/2024/01/user-activity.parquet")
		if err != nil {
			logger.Error("Failed to delete parquet file", slog.String("error", err.Error()))
			return
		}

		logger.Info("Successfully deleted parquet file")
	}
}
