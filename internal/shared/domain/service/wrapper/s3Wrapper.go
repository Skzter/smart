package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

const parquetFileExtension = ".parquet"

// S3Wrapper provides methods to interact with AWS S3 for parquet files
type S3Wrapper struct {
	client *s3.Client
	config entity.S3Config
	logger *slog.Logger
}

// NewS3Wrapper creates a new S3Wrapper instance
func NewS3Wrapper(logger *slog.Logger, config entity.S3Config) (*S3Wrapper, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, fmt.Errorf("logger cannot be nil: %w", err)
	}

	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	if config.Region == "" {
		config.Region = "us-east-1" // Default region
	}

	// Load AWS configuration
	var cfg aws.Config
	var err error

	if config.AccessKey != "" && config.SecretKey != "" {
		// Use provided credentials
		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithRegion(config.Region),
			awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     config.AccessKey,
					SecretAccessKey: config.SecretKey,
				}, nil
			})),
		)
	} else {
		// Use default credential chain (environment, instance role, etc.)
		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithRegion(config.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint if provided (for S3-compatible services)
	var s3Client *s3.Client
	if config.Endpoint != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(config.Endpoint)
			o.UsePathStyle = true // Required for some S3-compatible services
		})
	} else {
		s3Client = s3.NewFromConfig(cfg)
	}

	wrapper := &S3Wrapper{
		client: s3Client,
		config: config,
		logger: logger,
	}

	logger.Info("S3Wrapper initialized",
		slog.String("region", config.Region),
		slog.String("bucket", config.Bucket),
		slog.String("endpoint", config.Endpoint),
	)

	return wrapper, nil
}

// UploadParquetFile uploads a parquet file to S3
func (s *S3Wrapper) UploadParquetFile(ctx context.Context, key string, data []byte, metadata map[string]string) error {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return err
	}

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	// Ensure key ends with .parquet
	if len(key) < 8 || key[len(key)-8:] != parquetFileExtension {
		key += parquetFileExtension
	}

	// Prepare metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["Content-Type"] = "application/octet-stream"
	metadata["File-Type"] = "parquet"
	metadata["Upload-Time"] = time.Now().UTC().Format(time.RFC3339)

	s.logger.Info("Uploading parquet file to S3",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
		slog.Int("size_bytes", len(data)),
	)

	// Upload to S3
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/octet-stream"),
		Metadata:    metadata,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		s.logger.Error("Failed to upload parquet file",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to upload parquet file: %w", err)
	}

	s.logger.Info("Successfully uploaded parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	return nil
}

// DownloadParquetFile downloads a parquet file from S3
func (s *S3Wrapper) DownloadParquetFile(ctx context.Context, key string) ([]byte, map[string]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return nil, nil, err
	}

	if key == "" {
		return nil, nil, fmt.Errorf("key cannot be empty")
	}

	s.logger.Info("Downloading parquet file from S3",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	// Download from S3
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		s.logger.Error("Failed to download parquet file",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, nil, fmt.Errorf("failed to download parquet file: %w", err)
	}
	defer func() {
		if closeErr := result.Body.Close(); closeErr != nil {
			s.logger.Error("Failed to close response body",
				slog.String("key", key),
				slog.String("error", closeErr.Error()),
			)
		}
	}()

	// Read the data
	data, err := io.ReadAll(result.Body)
	if err != nil {
		s.logger.Error("Failed to read downloaded file",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, nil, fmt.Errorf("failed to read downloaded file: %w", err)
	}

	s.logger.Info("Successfully downloaded parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
		slog.Int("size_bytes", len(data)),
	)

	return data, result.Metadata, nil
}

// ListParquetFiles lists all parquet files in the bucket with optional prefix
func (s *S3Wrapper) ListParquetFiles(ctx context.Context, prefix string) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("Listing parquet files",
		slog.String("bucket", s.config.Bucket),
		slog.String("prefix", prefix),
	)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.config.Bucket),
	}

	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			s.logger.Error("Failed to list objects",
				slog.String("bucket", s.config.Bucket),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range output.Contents {
			key := aws.ToString(obj.Key)
			// Filter for parquet files
			if len(key) >= 8 && key[len(key)-8:] == parquetFileExtension {
				keys = append(keys, key)
			}
		}
	}

	s.logger.Info("Listed parquet files",
		slog.String("bucket", s.config.Bucket),
		slog.Int("count", len(keys)),
	)

	return keys, nil
}

// DeleteParquetFile deletes a parquet file from S3
func (s *S3Wrapper) DeleteParquetFile(ctx context.Context, key string) error {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return err
	}

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.logger.Info("Deleting parquet file from S3",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		s.logger.Error("Failed to delete parquet file",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to delete parquet file: %w", err)
	}

	s.logger.Info("Successfully deleted parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	return nil
}

// FileExists checks if a parquet file exists in S3
func (s *S3Wrapper) FileExists(ctx context.Context, key string) (bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return false, err
	}

	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetFileSize returns the size of a parquet file in S3
func (s *S3Wrapper) GetFileSize(ctx context.Context, key string) (int64, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("context cannot be nil", slog.String("error", err.Error()))
		return 0, err
	}

	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return aws.ToInt64(result.ContentLength), nil
}
