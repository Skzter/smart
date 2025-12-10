//go:build integration

package service

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

func TestSaveTestcaseIntegration(t *testing.T) {
	setup := setupTestService(t)

	testID := "integration-test-id"
	testcase := &entity.TestCase{
		TestID: testID,
		TestCode: entity.TestCode{
			Code: "best auto-playwrigth test in existens",
		},
		Status: entity.TestStatusPassed,
	}
	userId := "integration-user"

	key, err := setup.service.SaveTestcase(setup.ctx, testcase, userId)
	require.NoError(t, err)

	loaded, err := setup.repo.Read(setup.ctx, key)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, testcase.TestID, loaded.TestID)
	assert.Equal(t, testcase.TestCode.Code, loaded.TestCode.Code)
}

const (
	testBucketName = "test-bucket"
	testAccessKey  = "minioadmin"
	testSecretKey  = "minioadmin"
)

type testServiceSetup struct {
	service TestcaseStorageService
	repo    repository.TestcaseStorageRepository
	ctx     context.Context
}

func setupTestService(t *testing.T) *testServiceSetup {
	minioContainer, cleanup := setupMinIOContainer(t)
	t.Cleanup(cleanup)

	ctx := context.Background()
	logger := slog.Default()
	tracer := otel.Tracer("test")

	endpoint, err := minioContainer.ConnectionString(ctx)
	require.NoError(t, err)
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	s3Config := wrapperEntity.S3Config{
		Region:    "us-east-1",
		Bucket:    testBucketName,
		AccessKey: testAccessKey,
		SecretKey: testSecretKey,
		Endpoint:  endpoint,
	}

	createBucket(t, minioContainer)

	s3Wrapper, err := wrapperService.NewS3Wrapper(logger, s3Config, tracer)
	require.NoError(t, err)
	require.NotNil(t, s3Wrapper)

	parquetConfig := wrapperService.DefaultParquetConfig()
	parquetWrapper, err := wrapperService.NewParquetWrapper[entity.TestCase](logger, parquetConfig, tracer)
	require.NoError(t, err)
	require.NotNil(t, parquetWrapper)

	s3Repo, err := repository.NewTestcaseStorageRepository(logger, s3Wrapper, parquetWrapper, "prefixTest/", tracer)
	require.NoError(t, err)
	require.NotNil(t, s3Repo)

	service, err := NewTestcaseStorageService(logger, s3Repo, tracer)
	require.NoError(t, err)
	require.NotNil(t, service)

	return &testServiceSetup{
		service: service,
		repo:    s3Repo,
		ctx:     ctx,
	}
}

// createBucket creates the test bucket using the MinIO container
func createBucket(t *testing.T, minioContainer *minio.MinioContainer) {
	ctx := context.Background()

	endpoint, err := minioContainer.ConnectionString(ctx)
	require.NoError(t, err)

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

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

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(testBucketName),
	})

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
