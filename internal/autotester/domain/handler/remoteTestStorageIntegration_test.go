//go:build integration

package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/minio"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
	wrapperEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity/wrapper"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint:funlen
func TestHandleGetRemoteTestcaseIntegration(t *testing.T) {
	router, ids, testcasesMetadata := setupTestHandler(t)

	oneDayAgo := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	oneDayFuture := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
		expectedReturn map[string]entity.TestcaseMetadata
	}{
		{
			name:           "no filter",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedReturn: testcasesMetadata,
		},
		{
			name:           "filter by user1",
			queryParams:    "?author=" + ids["user1"],
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedReturn: map[string]entity.TestcaseMetadata{
				ids["test1"]: testcasesMetadata[ids["test1"]],
				ids["test2"]: testcasesMetadata[ids["test2"]],
			},
		},
		{
			name:           "filter by user2",
			queryParams:    "?author=" + ids["user2"],
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedReturn: map[string]entity.TestcaseMetadata{
				ids["test3"]: testcasesMetadata[ids["test3"]],
				ids["test4"]: testcasesMetadata[ids["test4"]],
			},
		},
		{
			name:           "filter by testcaseId test1",
			queryParams:    "?testcaseId=" + ids["test1"],
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedReturn: map[string]entity.TestcaseMetadata{
				ids["test1"]: testcasesMetadata[ids["test1"]],
			},
		},
		{
			name:           "filter by createdAfter one day ago",
			queryParams:    "?createdAfter=" + oneDayAgo,
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedReturn: testcasesMetadata,
		},
		{
			name:           "filter by createdBefore one day future",
			queryParams:    "?createdBefore=" + oneDayFuture,
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedReturn: testcasesMetadata,
		},
		{
			name:           "filter by createdBefore one day ago",
			queryParams:    "?createdBefore=" + oneDayAgo,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedReturn: map[string]entity.TestcaseMetadata{},
		},
		{
			name:           "filter by createdAfter one day future",
			queryParams:    "?createdAfter=" + oneDayFuture,
			expectedStatus: http.StatusOK,
			expectedCount:  0,
			expectedReturn: map[string]entity.TestcaseMetadata{},
		},
		{
			name:           "invalid createdAfter",
			queryParams:    "?createdAfter=not-a-date",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "invalid createdBefore",
			queryParams:    "?createdBefore=not-a-date",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "invalid limit",
			queryParams:    "?limit=not-a-number",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "invalid offset",
			queryParams:    "?offset=not-a-number",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "negativ limit",
			queryParams:    "?limit=-1",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "negativ offset",
			queryParams:    "?offset=-1",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "author not a uuid",
			queryParams:    "?author=not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "testcaseId not a uuid",
			queryParams:    "?testcaseId=not-a-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
			expectedReturn: nil,
		},
		{
			name:           "filter by author and testcaseId",
			queryParams:    "?author=" + ids["user1"] + "&testcaseId=" + ids["test1"],
			expectedStatus: http.StatusOK,
			expectedCount:  1,
			expectedReturn: map[string]entity.TestcaseMetadata{
				ids["test1"]: testcasesMetadata[ids["test1"]],
			},
		},
		{
			name:           "filter by author and createdAfter",
			queryParams:    "?author=" + ids["user1"] + "&createdAfter=" + oneDayAgo,
			expectedStatus: http.StatusOK,
			expectedCount:  2,
			expectedReturn: map[string]entity.TestcaseMetadata{
				ids["test1"]: testcasesMetadata[ids["test1"]],
				ids["test2"]: testcasesMetadata[ids["test2"]],
			},
		},
		{
			name:           "filter by createdAfter and createdBefore",
			queryParams:    "?createdAfter=" + oneDayAgo + "&createdBefore=" + oneDayFuture,
			expectedStatus: http.StatusOK,
			expectedCount:  4,
			expectedReturn: testcasesMetadata,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/tests"+test.queryParams, nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			assert.Equal(t, test.expectedStatus, resp.Code, "HTTP status does not match expected")
			if resp.Code == http.StatusOK {
				var result []*entity.TestcaseMetadata
				err := json.Unmarshal(resp.Body.Bytes(), &result)
				require.NoError(t, err, "failed to parse response body")
				assert.Equal(t, test.expectedCount, len(result), "number of testcases does not match expected count")
				for _, metadata := range result {
					expectedMetadata, exists := test.expectedReturn[metadata.TestcaseId]
					require.True(t, exists, "testcase %s not found in expected return", metadata.TestcaseId)
					assert.Equal(t, expectedMetadata.Key, metadata.Key, "key does not match for testcase %s", metadata.TestcaseId)
					assert.Equal(t, expectedMetadata.TestcaseId, metadata.TestcaseId, "testcaseId does not match")
					assert.Equal(t, expectedMetadata.Author, metadata.Author, "author does not match for testcase %s", metadata.TestcaseId)
				}
			}
		})
	}
}

func setupTestHandler(t *testing.T) (*gin.Engine, map[string]string, map[string]entity.TestcaseMetadata) {
	setup := setupTestService(t)

	user1 := uuid.New().String()
	user2 := uuid.New().String()
	test1 := uuid.New().String()
	test2 := uuid.New().String()
	test3 := uuid.New().String()
	test4 := uuid.New().String()

	testcases := []struct {
		testcase *entity.TestCase
		userId   string
	}{
		{
			testcase: &entity.TestCase{
				TestID: test1,
				TestCode: entity.TestCode{
					Code: "test login functionality",
				},
				Status: entity.TestStatusPassed,
			},
			userId: user1,
		},
		{
			testcase: &entity.TestCase{
				TestID: test2,
				TestCode: entity.TestCode{
					Code: "test logout functionality",
				},
				Status: entity.TestStatusPassed,
			},
			userId: user1,
		},
		{
			testcase: &entity.TestCase{
				TestID: test3,
				TestCode: entity.TestCode{
					Code: "test integration",
				},
				Status: entity.TestStatusPassed,
			},
			userId: user2,
		},
		{
			testcase: &entity.TestCase{
				TestID: test4,
				TestCode: entity.TestCode{
					Code: "test performance",
				},
				Status: entity.TestStatusPassed,
			},
			userId: user2,
		},
	}

	testcaseMetadata := make(map[string]entity.TestcaseMetadata)
	for _, testcase := range testcases {
		key, err := setup.service.SaveTestcase(setup.ctx, testcase.testcase, testcase.userId)
		require.NoError(t, err, "failed to save testcase: %s", testcase.testcase.TestID)

		testcaseMetadata[testcase.testcase.TestID] = entity.TestcaseMetadata{
			Key:        key,
			TestcaseId: testcase.testcase.TestID,
			Author:     testcase.userId,
			Name:       "",
		}
	}

	cfg, _ := config.LoadConfig()

	controller := &AutotesterController{
		remoteTestcaseStorageService: setup.service,
		logger:                       slog.Default(),
		config:                       cfg,
	}

	router := gin.Default()
	router.GET("/api/v1/tests", controller.HandleGetRemoteTestcase)

	controllerTestIDs := map[string]string{
		"user1": user1,
		"user2": user2,
		"test1": test1,
		"test2": test2,
		"test3": test3,
		"test4": test4,
	}

	return router, controllerTestIDs, testcaseMetadata
}

const (
	testBucketName = "test-bucket"
	testAccessKey  = "minioadmin"
	testSecretKey  = "minioadmin"
)

type testServiceSetup struct {
	service service.TestcaseStorageService
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

	service, err := service.NewTestcaseStorageService(logger, s3Repo, tracer)
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
