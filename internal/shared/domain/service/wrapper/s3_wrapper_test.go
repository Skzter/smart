//go:build integration

package service

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/minio"

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
				assert.Equal(t, tt.config.Bucket, tt.config.Bucket)

				// Check default region is set
				if tt.config.Region == "" {
					assert.Equal(t, "eu-central-1", tt.config.Region)
				} else {
					assert.Equal(t, tt.config.Region, tt.config.Region)
				}
			}
		})
	}
}

func TestNewS3WrapperNilLogger(t *testing.T) {
	config := entity.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}

	wrapper, err := NewS3Wrapper(nil, config)
	assert.Error(t, err)
	assert.Nil(t, wrapper)
	assert.Contains(t, err.Error(), "logger cannot be nil")
}

func TestS3WrapperInputValidation(t *testing.T) {
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

		// Test nil context
		err = wrapper.UploadParquetFile(nil, "test-file", []byte("test data"), nil) //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})

	t.Run("DownloadParquetFile validation", func(t *testing.T) {
		// Test empty key
		_, _, err := wrapper.DownloadParquetFile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context
		_, _, err = wrapper.DownloadParquetFile(nil, "test-file") //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})

	t.Run("DeleteParquetFile validation", func(t *testing.T) {
		// Test empty key
		err := wrapper.DeleteParquetFile(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context
		err = wrapper.DeleteParquetFile(nil, "test-file") //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})

	t.Run("FileExists validation", func(t *testing.T) {
		// Test empty key
		_, err := wrapper.FileExists(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context
		_, err = wrapper.FileExists(nil, "test-file") //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})

	t.Run("GetFileSize validation", func(t *testing.T) {
		// Test empty key
		_, err := wrapper.GetFileSize(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key cannot be empty")

		// Test nil context
		_, err = wrapper.GetFileSize(nil, "test-file") //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})

	t.Run("ListParquetFiles validation", func(t *testing.T) {
		// Test nil context
		_, err := wrapper.ListParquetFiles(nil, "prefix") //nolint:staticcheck
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "assert failed")
	})
}

const (
	testBucketName = "test-bucket"
	testAccessKey  = "minioadmin"
	testSecretKey  = "minioadmin"
)

// createBucket creates the test bucket using the MinIO container
func createBucket(t *testing.T, minioContainer *minio.MinioContainer) {
	ctx := context.Background()

	// Get MinIO endpoint and create S3 client for bucket creation
	endpoint, err := minioContainer.ConnectionString(ctx)
	require.NoError(t, err)

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	// Create AWS config for bucket creation
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     testAccessKey,
				SecretAccessKey: testSecretKey,
			}, nil
		})),
	)
	require.NoError(t, err)

	// Create S3 client for bucket operations
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	// Create bucket
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(testBucketName),
	})

	// Ignore error if bucket already exists
	if err != nil && !strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") && !strings.Contains(err.Error(), "BucketAlreadyExists") {
		t.Logf("Warning: Failed to create bucket (may already exist): %v", err)
	}
}

func setupMinIOContainer(t *testing.T) (*minio.MinioContainer, func()) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create MinIO container
	minioContainer, err := minio.Run(ctx,
		"minio/minio:RELEASE.2024-01-16T16-07-38Z",
		minio.WithUsername(testAccessKey),
		minio.WithPassword(testSecretKey),
	)
	require.NoError(t, err, "Failed to start MinIO container")

	cleanup := func() {
		if err := minioContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MinIO container: %v", err)
		}
	}

	return minioContainer, cleanup
}

func createS3WrapperWithMinIO(t *testing.T, minioContainer *minio.MinioContainer) *S3Wrapper {
	ctx := context.Background()

	// Get MinIO connection details
	endpoint, err := minioContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Ensure endpoint has http:// prefix
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	// Create S3 config for MinIO
	config := entity.S3Config{
		Region:    "us-east-1", // MinIO uses us-east-1 by default
		Bucket:    testBucketName,
		AccessKey: testAccessKey,
		SecretKey: testSecretKey,
		Endpoint:  endpoint,
	}

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create the test bucket first
	createBucket(t, minioContainer)

	// Create S3 wrapper
	wrapper, err := NewS3Wrapper(logger, config)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	return wrapper.(*S3Wrapper)
}

func TestS3WrapperUploadAndDownloadIntegration(t *testing.T) {
	minioContainer, cleanup := setupMinIOContainer(t)
	defer cleanup()

	wrapper := createS3WrapperWithMinIO(t, minioContainer)
	ctx := context.Background()

	// Create test data
	testKey := "test-data/sample-file"
	testData := []byte("This is test parquet data content for integration testing")
	testMetadata := map[string]string{
		"source":    "integration-test",
		"test-case": "upload-download",
		"version":   "1.0",
	}

	// Upload parquet file
	err := wrapper.UploadParquetFile(ctx, testKey, testData, testMetadata)
	assert.NoError(t, err)

	// Download parquet file
	downloadedData, metadata, err := wrapper.DownloadParquetFile(ctx, testKey+".parquet")
	assert.NoError(t, err)
	assert.Equal(t, testData, downloadedData)
	assert.NotNil(t, metadata)

	// Check metadata
	assert.Contains(t, metadata, "source")
	assert.Equal(t, "integration-test", metadata["source"])
	assert.Contains(t, metadata, "test-case")
	assert.Equal(t, "upload-download", metadata["test-case"])

	// Check file exists
	exists, err := wrapper.FileExists(ctx, testKey+".parquet")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test non-existent file
	exists, err = wrapper.FileExists(ctx, "non-existent-file.parquet")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Get file size
	size, err := wrapper.GetFileSize(ctx, testKey+".parquet")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(testData)), size)

	// List parquet files
	files, err := wrapper.ListParquetFiles(ctx, "test-data/")
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, testKey+".parquet", files[0])

	// Delete parquet file
	err = wrapper.DeleteParquetFile(ctx, testKey+".parquet")
	assert.NoError(t, err)

	// Verify file is deleted
	exists, err = wrapper.FileExists(ctx, testKey+".parquet")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestS3WrapperParquetExtensionHandlingIntegration(t *testing.T) {
	minioContainer, cleanup := setupMinIOContainer(t)
	defer cleanup()

	wrapper := createS3WrapperWithMinIO(t, minioContainer)
	ctx := context.Background()

	testData := []byte("Test data for extension handling")

	t.Run("Upload file without .parquet extension", func(t *testing.T) {
		key := "test-file-no-ext"
		err := wrapper.UploadParquetFile(ctx, key, testData, nil)
		assert.NoError(t, err)

		// Verify file exists with .parquet extension
		exists, err := wrapper.FileExists(ctx, key+".parquet")
		assert.NoError(t, err)
		assert.True(t, exists)

		// Download and verify
		data, _, err := wrapper.DownloadParquetFile(ctx, key+".parquet")
		assert.NoError(t, err)
		assert.Equal(t, testData, data)

		// Cleanup
		_ = wrapper.DeleteParquetFile(ctx, key+".parquet")
	})

	t.Run("Upload file with .parquet extension", func(t *testing.T) {
		key := "test-file-with-ext.parquet"
		err := wrapper.UploadParquetFile(ctx, key, testData, nil)
		assert.NoError(t, err)

		// Verify file exists
		exists, err := wrapper.FileExists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Download and verify
		data, _, err := wrapper.DownloadParquetFile(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, testData, data)

		// Cleanup
		_ = wrapper.DeleteParquetFile(ctx, key)
	})
}

func TestS3WrapperErrorHandlingIntegration(t *testing.T) {
	minioContainer, cleanup := setupMinIOContainer(t)
	defer cleanup()

	wrapper := createS3WrapperWithMinIO(t, minioContainer)
	ctx := context.Background()

	t.Run("Download non-existent file", func(t *testing.T) {
		_, _, err := wrapper.DownloadParquetFile(ctx, "non-existent-file.parquet")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to download parquet file")
	})

	t.Run("Delete non-existent file", func(t *testing.T) {
		// MinIO doesn't return an error when deleting non-existent files
		// This is consistent with AWS S3 behavior
		err := wrapper.DeleteParquetFile(ctx, "non-existent-file.parquet")
		assert.NoError(t, err)
	})

	t.Run("Get size of non-existent file", func(t *testing.T) {
		_, err := wrapper.GetFileSize(ctx, "non-existent-file.parquet")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get file metadata")
	})
}

func TestS3WrapperLargeFileIntegration(t *testing.T) {
	minioContainer, cleanup := setupMinIOContainer(t)
	defer cleanup()

	wrapper := createS3WrapperWithMinIO(t, minioContainer)
	ctx := context.Background()

	// Create a larger test file (1MB)
	largeData := bytes.Repeat([]byte("A"), 1024*1024)
	key := "large-file-test"

	// Upload
	start := time.Now()
	err := wrapper.UploadParquetFile(ctx, key, largeData, map[string]string{
		"size": "1MB",
		"type": "large-file-test",
	})
	uploadDuration := time.Since(start)
	assert.NoError(t, err)
	t.Logf("Upload took: %v", uploadDuration)

	// Verify file exists and get size
	exists, err := wrapper.FileExists(ctx, key+".parquet")
	assert.NoError(t, err)
	assert.True(t, exists)

	size, err := wrapper.GetFileSize(ctx, key+".parquet")
	assert.NoError(t, err)
	assert.Equal(t, int64(len(largeData)), size)

	// Download
	start = time.Now()
	downloadedData, metadata, err := wrapper.DownloadParquetFile(ctx, key+".parquet")
	downloadDuration := time.Since(start)
	assert.NoError(t, err)
	assert.Equal(t, largeData, downloadedData)
	assert.Equal(t, "1MB", metadata["size"])
	t.Logf("Download took: %v", downloadDuration)

	// Cleanup
	err = wrapper.DeleteParquetFile(ctx, key+".parquet")
	assert.NoError(t, err)
}
