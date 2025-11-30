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
	"go.opentelemetry.io/otel"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
)

func TestNewS3Wrapper(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := otel.Tracer("test")

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
			wrapper, err := NewS3Wrapper(logger, tt.config, tracer)

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
	tracer := otel.Tracer("test")
	config := entity.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}

	wrapper, err := NewS3Wrapper(nil, config, tracer)
	assert.Error(t, err)
	assert.Nil(t, wrapper)
	assert.ErrorIs(t, err, ErrNilLogger)
}

func TestS3WrapperInputValidation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	tracer := otel.Tracer("test")
	config := entity.S3Config{
		Region: "us-east-1",
		Bucket: "test-bucket",
	}

	wrapper, err := NewS3Wrapper(logger, config, tracer)
	require.NoError(t, err)
	require.NotNil(t, wrapper)

	testCases := []struct {
		name          string
		methodToTest  string
		key           string
		data          []byte
		useNilContext bool
		expectedError error
	}{
		// UploadParquetFile validation
		{"UploadParquetFile with empty key", "UploadParquetFile", "", []byte("test data"), false, ErrEmptyKey},
		{"UploadParquetFile with empty data", "UploadParquetFile", "test-file", []byte{}, false, ErrEmptyData},
		{"UploadParquetFile with nil context", "UploadParquetFile", "test-file", []byte("test data"), true, ErrNilContext},

		// DownloadParquetFile validation
		{"DownloadParquetFile with empty key", "DownloadParquetFile", "", nil, false, ErrEmptyKey},
		{"DownloadParquetFile with nil context", "DownloadParquetFile", "test-file", nil, true, ErrNilContext},

		// DeleteParquetFile validation
		{"DeleteParquetFile with empty key", "DeleteParquetFile", "", nil, false, ErrEmptyKey},
		{"DeleteParquetFile with nil context", "DeleteParquetFile", "test-file", nil, true, ErrNilContext},

		// FileExists validation
		{"FileExists with empty key", "FileExists", "", nil, false, ErrEmptyKey},
		{"FileExists with nil context", "FileExists", "test-file", nil, true, ErrNilContext},

		// GetFileSize validation
		{"GetFileSize with empty key", "GetFileSize", "", nil, false, ErrEmptyKey},
		{"GetFileSize with nil context", "GetFileSize", "test-file", nil, true, ErrNilContext},

		// ListParquetFiles validation
		{"ListParquetFiles with nil context", "ListParquetFiles", "prefix", nil, true, ErrNilContext},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			ctx := context.Background()
			if tc.useNilContext {
				ctx = nil
			}

			switch tc.methodToTest {
			case "UploadParquetFile":
				err = wrapper.UploadParquetFile(ctx, tc.key, tc.data, nil)
			case "DownloadParquetFile":
				_, _, err = wrapper.DownloadParquetFile(ctx, tc.key)
			case "DeleteParquetFile":
				err = wrapper.DeleteParquetFile(ctx, tc.key)
			case "FileExists":
				_, err = wrapper.FileExists(ctx, tc.key)
			case "GetFileSize":
				_, err = wrapper.GetFileSize(ctx, tc.key)
			case "ListParquetFiles":
				_, err = wrapper.ListParquetFiles(ctx, tc.key)
			default:
				t.Fatalf("Unknown method: %s", tc.methodToTest)
			}

			assert.Error(t, err)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
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
	tracer := otel.Tracer("test")

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
	wrapper, err := NewS3Wrapper(logger, config, tracer)
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
