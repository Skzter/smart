//go:build integration

package service

import (
	"context"
	"log/slog"
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

// nolint:funlen
func TestReadAllMetadataWithFilterIntegration(t *testing.T) {
	setup := setupTestService(t)

	testcases := []struct {
		testcase *entity.TestCase
		userId   string
	}{
		{
			testcase: &entity.TestCase{
				TestID: "test1",
				TestCode: entity.TestCode{
					Code: "test login functionality",
				},
				Status: entity.TestStatusPassed,
			},
			userId: "user1",
		},
		{
			testcase: &entity.TestCase{
				TestID: "test2",
				TestCode: entity.TestCode{
					Code: "test logout functionality",
				},
				Status: entity.TestStatusPassed,
			},
			userId: "user1",
		},
		{
			testcase: &entity.TestCase{
				TestID: "test3",
				TestCode: entity.TestCode{
					Code: "test integration",
				},
				Status: entity.TestStatusPassed,
			},
			userId: "user2",
		},
		{
			testcase: &entity.TestCase{
				TestID: "test4",
				TestCode: entity.TestCode{
					Code: "test performance",
				},
				Status: entity.TestStatusPassed,
			},
			userId: "user2",
		},
	}

	for _, tc := range testcases {
		_, err := setup.service.SaveTestcase(setup.ctx, tc.testcase, tc.userId)
		require.NoError(t, err, "failed to save testcase: %s", tc.testcase.TestID)
	}

	oneDayAgo := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	oneDayFuture := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	tests := []struct {
		name          string
		filter        *entity.GetRemoteTestcaseRequest
		expectedCount int
	}{
		{
			name:          "no filter",
			filter:        &entity.GetRemoteTestcaseRequest{},
			expectedCount: 4,
		},
		{
			name:          "filter by user1",
			filter:        &entity.GetRemoteTestcaseRequest{Author: "user1"},
			expectedCount: 2,
		},
		{
			name:          "filter by user2",
			filter:        &entity.GetRemoteTestcaseRequest{Author: "user2"},
			expectedCount: 2,
		},
		{
			name:          "filter by testcaseId test1",
			filter:        &entity.GetRemoteTestcaseRequest{TestcaseId: "test1"},
			expectedCount: 1,
		},
		{
			name:          "filter by testcaseId test2",
			filter:        &entity.GetRemoteTestcaseRequest{TestcaseId: "test2"},
			expectedCount: 1,
		},
		{
			name:          "filter by createdAfter one day ago - all pass",
			filter:        &entity.GetRemoteTestcaseRequest{CreatedAfter: oneDayAgo},
			expectedCount: 4,
		},
		{
			name:          "filter by createdBefore one day future - all pass",
			filter:        &entity.GetRemoteTestcaseRequest{CreatedBefore: oneDayFuture},
			expectedCount: 4,
		},
		{
			name:          "filter by createdBefore one day ago - none pass",
			filter:        &entity.GetRemoteTestcaseRequest{CreatedBefore: oneDayAgo},
			expectedCount: 0,
		},
		{
			name:          "filter by createdAfter one day future - none pass",
			filter:        &entity.GetRemoteTestcaseRequest{CreatedAfter: oneDayFuture},
			expectedCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := setup.service.ReadAllMetadataWithFilter(setup.ctx, test.filter)
			require.NoError(t, err)
			assert.Len(t, data, test.expectedCount, "unexpected number of results")
		})
	}
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
