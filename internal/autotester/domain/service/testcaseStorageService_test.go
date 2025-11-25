package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/minio"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint: dupl
func TestNewTestcaseStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := &mocks.MockTestCaseStorageRepository{}
	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.TestCaseStorageRepository
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			repo:    mockRepo,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			repo:    mockRepo,
			wantErr: true,
		},
		{
			name:    "nil repo",
			logger:  logger,
			repo:    nil,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewTestcaseStorageService(test.logger, test.repo)
			if (err != nil) != test.wantErr {
				t.Errorf("NewTestcaseStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && svc == nil {
				t.Errorf("NewTestcaseStorageService() returned nil service")
			}
		})
	}
}

// nolint: dupl
func TestSaveTestCase(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		context   context.Context
		testCase  *entity.TestCase
		userId    string
		createErr error
		wantErr   bool
	}{
		{
			name:      "success",
			context:   context.Background(),
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: nil,
			wantErr:   false,
		},
		{
			name:      "nil context",
			context:   nil,
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: nil,
			wantErr:   true,
		},
		{
			name:      "repo returns error",
			context:   context.Background(),
			testCase:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:    "valid user",
			createErr: errors.New("repo error"),
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := &mocks.MockTestCaseStorageRepository{}
			var key string
			if test.createErr == nil {
				key = "dummy-key"
			}
			mockRepo.EXPECT().Create(test.context, test.testCase, test.userId).Return(key, test.createErr)

			svc, err := NewTestcaseStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("unexpected error creating service: %v", err)
			}
			_, err = svc.SaveTestcase(test.context, test.testCase, test.userId)
			if (err != nil) != test.wantErr {
				t.Errorf("SaveTestCase() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestSaveTestcaseIntegration(t *testing.T) {
	minioContainer, cleanup := setupMinIOContainer(t)
	defer cleanup()

	ctx := context.Background()
	logger := slog.Default()

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

	s3Wrapper, err := wrapperService.NewS3Wrapper(logger, s3Config)
	require.NoError(t, err)
	require.NotNil(t, s3Wrapper)

	parquetConfig := wrapperService.DefaultParquetConfig()
	parquetWrapper, err := wrapperService.NewParquetWrapper[entity.TestCase](logger, parquetConfig)
	require.NoError(t, err)
	require.NotNil(t, parquetWrapper)

	s3Repo, err := repository.NewTestCaseStorageRepository(logger, s3Wrapper, parquetWrapper)
	require.NoError(t, err)
	require.NotNil(t, s3Repo)

	service, err := NewTestcaseStorageService(logger, s3Repo)
	require.NoError(t, err)
	require.NotNil(t, service)

	testID := "integration-test-id"
	testcase := &entity.TestCase{
		TestID: testID,
		TestCode: entity.TestCode{
			Code: "best auto-playwrigth test in existens",
		},
		Status: entity.TestStatusPassed,
	}
	userId := "integration-user"

	key, err := service.SaveTestcase(ctx, testcase, userId)
	require.NoError(t, err)

	loaded, err := s3Repo.Read(ctx, key)
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
