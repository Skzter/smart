package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

// nolint: dupl
func TestNewTestcaseStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockTestcaseStorageRepository(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.TestcaseStorageRepository
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
			svc, err := NewTestcaseStorageService(test.logger, test.repo, tracer)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, svc, "service should be nil on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, svc, "service should not be nil")
			}
		})
	}
}

// nolint: dupl
func TestSaveTestcase(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		context   context.Context
		testCase  *entity.TestCase
		userId    string
		wantErr   bool
		setupMock func(*mocks.MockTestcaseStorageRepository)
	}{
		{
			name:     "success",
			context:  ctx,
			testCase: &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:   "valid user",
			wantErr:  false,
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().Create(mock.Anything, mock.Anything, "valid user").Return("dummy-key", nil)
			},
		},
		{
			name:     "nil context",
			context:  nil,
			testCase: &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:   "valid user",
			wantErr:  true,
		},
		{
			name:     "repo returns error",
			context:  ctx,
			testCase: &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:   "valid user",
			wantErr:  true,
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().Create(mock.Anything, mock.Anything, "valid user").Return("", errors.New("repo error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseStorageRepository(t)

			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewTestcaseStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			_, err = svc.SaveTestcase(test.context, test.testCase, test.userId)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestReadAllMetadataWithFilter(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	tracer := otel.Tracer("test")

	metadata1 := &entity.TestcaseMetadata{
		Key:     "test-1",
		Author:  "user1",
		Created: "1000",
		Updated: "1500",
		Name:    "Test Case 1",
	}
	metadata2 := &entity.TestcaseMetadata{
		Key:     "test-2",
		Author:  "user2",
		Created: "2000",
		Updated: "2500",
		Name:    "Test Case 2",
	}
	allMetadata := []*entity.TestcaseMetadata{metadata1, metadata2}

	tests := []struct {
		name             string
		context          context.Context
		filter           *entity.GetRemoteTestcaseRequest
		wantErr          bool
		expectedResponse []*entity.TestcaseMetadata
		setupMock        func(*mocks.MockTestcaseStorageRepository)
	}{
		{
			name:             "success - no filter",
			context:          ctx,
			filter:           &entity.GetRemoteTestcaseRequest{},
			wantErr:          false,
			expectedResponse: allMetadata,
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().ReadAllMetadata(mock.Anything).Return(allMetadata, nil)
			},
		},
		{
			name:             "success - with filter",
			context:          ctx,
			filter:           &entity.GetRemoteTestcaseRequest{Author: "user1"},
			wantErr:          false,
			expectedResponse: []*entity.TestcaseMetadata{metadata1},
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().ReadAllMetadata(mock.Anything).Return(allMetadata, nil)
			},
		},
		{
			name:             "success - no results",
			context:          ctx,
			filter:           &entity.GetRemoteTestcaseRequest{Author: "user3"},
			wantErr:          false,
			expectedResponse: []*entity.TestcaseMetadata{},
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().ReadAllMetadata(mock.Anything).Return(allMetadata, nil)
			},
		},
		{
			name:             "nil context",
			context:          nil,
			filter:           &entity.GetRemoteTestcaseRequest{},
			wantErr:          true,
			expectedResponse: nil,
		},
		{
			name:             "repo returns error",
			context:          ctx,
			filter:           &entity.GetRemoteTestcaseRequest{},
			wantErr:          true,
			expectedResponse: nil,
			setupMock: func(mockRepo *mocks.MockTestcaseStorageRepository) {
				mockRepo.EXPECT().ReadAllMetadata(mock.Anything).Return(nil, errors.New("db error"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseStorageRepository(t)
			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewTestcaseStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			data, err := svc.ReadAllMetadataWithFilter(test.context, test.filter)

			if test.wantErr {
				assert.Error(t, err, "expected an error but got none")
				assert.Nil(t, data, "expected nil data when error occurs")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, test.expectedResponse, data, "response should match expected metadata")
			}
		})
	}
}

// nolint:funlen
func TestPassesAllFilters(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	mockRepo := mocks.NewMockTestcaseStorageRepository(t)

	svc := &testcaseStorageService{
		logger: logger,
		repo:   mockRepo,
		tracer: tracer,
	}

	metadata := &entity.TestcaseMetadata{
		Key:     "test-1",
		Author:  "user1",
		Created: "1609459200", // 2021-01-01T00:00:00Z
		Updated: "1609459500",
		Name:    "Test Case 1",
	}

	tests := []struct {
		name     string
		metadata *entity.TestcaseMetadata
		filter   *entity.GetRemoteTestcaseRequest
		wantPass bool
	}{
		{
			name:     "no filter - passes",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{},
			wantPass: true,
		},
		{
			name:     "author matches - passes",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{Author: "user1"},
			wantPass: true,
		},
		{
			name:     "author not matches - fails",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{Author: "user2"},
			wantPass: false,
		},
		{
			name:     "testcaseId in key - passes",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{TestcaseId: "test"},
			wantPass: true,
		},
		{
			name:     "testcaseId not in key - fails",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{TestcaseId: "integration"},
			wantPass: false,
		},
		// timestamp: 2021-01-01T00:00:00Z
		{
			name:     "createdAfter - timestamp after filter - passes",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{CreatedAfter: "2020-12-31T00:00:00Z"},
			wantPass: true,
		},
		{
			name:     "createdAfter - timestamp befor filter - fails",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{CreatedAfter: "2021-01-02T00:00:00Z"},
			wantPass: false,
		},
		{
			name:     "createdBefore - timestamp before filter - passes",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{CreatedBefore: "2021-01-02T00:00:00Z"},
			wantPass: true,
		},
		{
			name:     "createdBefore - timestamp after filter - fails",
			metadata: metadata,
			filter:   &entity.GetRemoteTestcaseRequest{CreatedBefore: "2020-12-31T00:00:00Z"},
			wantPass: false,
		},
		{
			name:     "combined filters all match - passes",
			metadata: metadata,
			filter: &entity.GetRemoteTestcaseRequest{
				Author:       "user1",
				TestcaseId:   "test",
				CreatedAfter: "1970-01-01T00:00:00Z",
			},
			wantPass: true,
		},
		{
			name:     "combined filters one not match - fails",
			metadata: metadata,
			filter: &entity.GetRemoteTestcaseRequest{
				Author:       "user1",
				TestcaseId:   "integration",
				CreatedAfter: "1970-01-01T00:00:00Z",
			},
			wantPass: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := svc.passesAllFilters(test.metadata, test.filter)
			assert.Equal(t, test.wantPass, result, "filter result should match expected")
		})
	}
}

func TestStringToInt64(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	mockRepo := mocks.NewMockTestcaseStorageRepository(t)

	svc := &testcaseStorageService{
		logger: logger,
		repo:   mockRepo,
		tracer: tracer,
	}

	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "valid timestamp string",
			input:    "1609459200",
			expected: 1609459200,
		},
		{
			name:     "invalid timestamp string",
			input:    "not-a-number",
			expected: 0,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := svc.stringToInt64(test.input)
			assert.Equal(t, test.expected, result, "stringToInt64 result should match expected")
		})
	}
}

func TestISO8601ToUnix(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")
	mockRepo := mocks.NewMockTestcaseStorageRepository(t)

	svc := &testcaseStorageService{
		logger: logger,
		repo:   mockRepo,
		tracer: tracer,
	}

	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "valid ISO 8601 timestamp",
			input:    "2021-01-01T00:00:00Z",
			expected: 1609459200,
		},
		{
			name:     "invalid ISO 8601 timestamp",
			input:    "not-a-date",
			expected: 0,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := svc.iso8601ToUnix(test.input)
			assert.Equal(t, test.expected, result, "iso8601ToUnix result should match expected")
		})
	}
}
