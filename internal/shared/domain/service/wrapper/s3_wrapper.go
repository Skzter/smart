package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

const parquetFileExtension = ".parquet"

var (
	// ErrNilLogger indicates that a required logger instance is nil.
	ErrNilLogger = errors.New("logger cannot be nil")

	// ErrNilContext indicates that a required context is nil.
	ErrNilContext = errors.New("context cannot be nil")

	// ErrEmptyKey indicates that a required key is empty.
	ErrEmptyKey = errors.New("key cannot be empty")

	// ErrEmptyData indicates that provided data is empty.
	ErrEmptyData = errors.New("data cannot be empty")

	// ErrEmptyBucket indicates that a required bucket name is empty.
	ErrEmptyBucket = errors.New("bucket cannot be empty")

	// ErrEmptyRegion indicates that a required region is empty.
	ErrEmptyRegion = errors.New("region cannot be empty")

	// ErrEmptyAccessKey indicates that the AWS access key is empty.
	ErrEmptyAccessKey = errors.New("access key cannot be empty")

	// ErrEmptySecretKey indicates that the AWS secret key is empty.
	ErrEmptySecretKey = errors.New("secret key cannot be empty")

	// ErrFailedToLoadAWSConfig indicates that loading the AWS configuration failed.
	ErrFailedToLoadAWSConfig = errors.New("failed to load AWS config")

	// ErrFailedToGetFileMetadata indicates that retrieving file metadata failed.
	ErrFailedToGetFileMetadata = errors.New("failed to get file metadata")
)

// S3StorageWrapper defines an interface for storing, retrieving and managing Parquet files in an S3-compatible storage service.
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

// S3Wrapper provides methods to interact with AWS S3 for parquet files
type S3Wrapper struct {
	client *s3.Client
	config entity.S3Config
	logger *slog.Logger
	tracer trace.Tracer
}

// NewS3Wrapper creates a new S3Wrapper instance
func NewS3Wrapper(logger *slog.Logger, config entity.S3Config, tracer trace.Tracer) (S3StorageWrapper, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, ErrNilLogger
	}

	if config.Bucket == "" {
		return nil, ErrEmptyBucket
	}

	if config.Region == "" {
		config.Region = "eu-central-1" // Default region
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
		return nil, ErrFailedToLoadAWSConfig
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
		tracer: tracer,
	}

	logger.Debug("S3Wrapper initialized",
		slog.String("region", config.Region),
		slog.String("bucket", config.Bucket),
		slog.String("endpoint", config.Endpoint),
	)

	return wrapper, nil
}

// UploadParquetFile uploads a parquet file to S3
func (s *S3Wrapper) UploadParquetFile(ctx context.Context, key string, data []byte, metadata map[string]string) error {
	if err := assert.NotNil(ctx); err != nil {
		return ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.UploadParquetFile")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return err
	}

	if len(data) == 0 {
		err := ErrEmptyData
		span.RecordError(err)
		span.SetStatus(codes.Error, "data validation failed")
		return err
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

	s.logger.Debug("Uploading parquet file to S3",
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 upload failed")
		s.logger.Error("Failed to upload parquet file",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to upload parquet file: %w", err)
	}

	span.AddEvent("Parquet file uploaded", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
	))
	span.SetStatus(codes.Ok, "")

	s.logger.Debug("Successfully uploaded parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	return nil
}

// DownloadParquetFile downloads a parquet file from S3
func (s *S3Wrapper) DownloadParquetFile(ctx context.Context, key string) ([]byte, map[string]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, nil, ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.DownloadParquetFile")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return nil, nil, err
	}

	s.logger.Debug("Downloading parquet file from S3",
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 download failed")
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
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read downloaded file")
		s.logger.Error("Failed to read downloaded file",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return nil, nil, fmt.Errorf("failed to read downloaded file: %w", err)
	}

	span.AddEvent("Parquet file downloaded", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
		attribute.Int("size_bytes", len(data)),
	))
	span.SetStatus(codes.Ok, "")

	s.logger.Debug("Successfully downloaded parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
		slog.Int("size_bytes", len(data)),
	)

	return data, result.Metadata, nil
}

// ListParquetFiles lists all parquet files in the bucket with optional prefix
func (s *S3Wrapper) ListParquetFiles(ctx context.Context, prefix string) ([]string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.ListParquetFiles")
	defer span.End()

	s.logger.Debug("Listing parquet files",
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
			span.RecordError(err)
			span.SetStatus(codes.Error, "S3 list objects failed")
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

	span.AddEvent("Parquet files listed", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.Int("count", len(keys)),
	))
	span.SetStatus(codes.Ok, "")

	s.logger.Debug("Listed parquet files",
		slog.String("bucket", s.config.Bucket),
		slog.Int("count", len(keys)),
	)

	return keys, nil
}

// DeleteParquetFile deletes a parquet file from S3
func (s *S3Wrapper) DeleteParquetFile(ctx context.Context, key string) error {
	if err := assert.NotNil(ctx); err != nil {
		return ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.DeleteParquetFile")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return err
	}

	s.logger.Debug("Deleting parquet file from S3",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 delete failed")
		s.logger.Error("Failed to delete parquet file",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to delete parquet file: %w", err)
	}

	span.AddEvent("Parquet file deleted", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
	))
	span.SetStatus(codes.Ok, "")

	s.logger.Debug("Successfully deleted parquet file",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
	)

	return nil
}

// FileExists checks if a parquet file exists in S3
func (s *S3Wrapper) FileExists(ctx context.Context, key string) (bool, error) {
	if err := assert.NotNil(ctx); err != nil {
		return false, ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.FileExists")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return false, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			span.SetStatus(codes.Ok, "")
			return false, nil
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 head object failed")
		return false, err
	}

	span.AddEvent("Parquet file exists", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
	))
	span.SetStatus(codes.Ok, "")

	return true, nil
}

// GetFileSize returns the size of a parquet file in S3
func (s *S3Wrapper) GetFileSize(ctx context.Context, key string) (int64, error) {
	if err := assert.NotNil(ctx); err != nil {
		return 0, ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.GetFileSize")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return 0, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get file metadata")
		return 0, ErrFailedToGetFileMetadata
	}

	span.AddEvent("Parquet file size retrieved", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
		attribute.Int64("size_bytes", aws.ToInt64(result.ContentLength)),
	))
	span.SetStatus(codes.Ok, "")

	return aws.ToInt64(result.ContentLength), nil
}

// UploadMediaFile uploads a media file to S3 (key with extension)
func (s *S3Wrapper) UploadMediaFile(ctx context.Context, key string, data []byte, metadata map[string]string) error {
	if err := assert.NotNil(ctx); err != nil {
		return ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.UploadMediaFile")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return err
	}

	if len(data) == 0 {
		err := ErrEmptyData
		span.RecordError(err)
		span.SetStatus(codes.Error, "data validation failed")
		return err
	}

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["Content-Type"] = "application/octet-stream"
	metadata["File-Type"] = "media"
	metadata["Upload-Time"] = time.Now().UTC().Format(time.RFC3339)

	s.logger.Debug("Uploading media file to S3",
		slog.String("bucket", s.config.Bucket),
		slog.String("key", key),
		slog.Int("size_bytes", len(data)),
	)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/octet-stream"),
		Metadata:    metadata,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 upload failed")
		return fmt.Errorf("failed to upload media file: %w", err)
	}

	span.AddEvent("Media file uploaded", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}

// GetMediaUrl returns the URL of a media file in S3 (key with extension)
func (s *S3Wrapper) GetMediaUrl(ctx context.Context, key string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return "", ErrNilContext
	}

	ctx, span := s.tracer.Start(ctx, "S3Wrapper.GetMediaUrl")
	defer span.End()

	if key == "" {
		err := ErrEmptyKey
		span.RecordError(err)
		span.SetStatus(codes.Error, "key validation failed")
		return "", err
	}

	// Check if file exists
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var notFound *types.NotFound
		errMsg := err.Error()
		isNotFound := errors.As(err, &notFound) ||
			strings.Contains(strings.ToLower(errMsg), "not found") ||
			strings.Contains(strings.ToLower(errMsg), "nosuchkey") ||
			strings.Contains(strings.ToLower(errMsg), "no such key")

		if isNotFound {
			err := fmt.Errorf("failed to get media url: file does not exist")
			span.RecordError(err)
			span.SetStatus(codes.Error, "file does not exist")
			s.logger.Error("Failed to get media url: file does not exist",
				slog.String("bucket", s.config.Bucket),
				slog.String("key", key),
			)
			return "", err
		}
		err := fmt.Errorf("failed to get media url: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "S3 head object failed")
		s.logger.Error("Failed to get media url",
			slog.String("bucket", s.config.Bucket),
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return "", err
	}

	span.AddEvent("Media URL generated", trace.WithAttributes(
		attribute.String("bucket", s.config.Bucket),
		attribute.String("key", key),
	))
	span.SetStatus(codes.Ok, "")

	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.config.Bucket, key), nil
}
