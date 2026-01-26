package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint:dupl
func TestCreate(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range testcaseCreateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.setupMock != nil {
				test.setupMock(mockS3, mockParquet)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
			}

			_, err := repo.Create(test.ctx, test.obj, test.userId)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func testcaseCreateTestCaseProvider() []struct {
	name        string
	obj         *entity.TestCase
	userId      string
	ctx         context.Context
	setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError bool
} {
	ctx := context.Background()
	return []struct {
		name        string
		obj         *entity.TestCase
		userId      string
		ctx         context.Context
		setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError bool
	}{
		{
			name:   "happy path",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
			ctx:    ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return([]byte("parquetdata"), nil)
				s3.EXPECT().UploadParquetFile(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "nil obj",
			obj:         nil,
			userId:      "valid user",
			ctx:         ctx,
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "nil ctx",
			obj:         &entity.TestCase{},
			userId:      "valid user",
			ctx:         nil,
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "validation fails",
			obj:         &entity.TestCase{},
			userId:      "valid user",
			ctx:         ctx,
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "userId validation fails",
			obj:         &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:      "",
			ctx:         ctx,
			setupMock:   nil,
			expectError: true,
		},
		{
			name:   "parquet error",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
			ctx:    ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return(nil, errors.New("parquet error"))
			},
			expectError: true,
		},
		{
			name:   "upload error",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
			ctx:    ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return([]byte("parquetdata"), nil)
				s3.EXPECT().UploadParquetFile(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("upload error"))
			},
			expectError: true,
		},
	}
}

// nolint:dupl
func TestReadRemote(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range readRemoteTestcaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.setupMock != nil {
				test.setupMock(mockS3, mockParquet)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
			}

			result, err := repo.Read(test.ctx, test.key)

			if test.expectError {
				assert.Error(t, err)
				if test.expectNilResult {
					assert.Nil(t, result)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func readRemoteTestcaseProvider() []struct {
	name            string
	key             string
	ctx             context.Context
	setupMock       func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError     bool
	expectNilResult bool
} {
	ctx := context.Background()
	return []struct {
		name            string
		key             string
		ctx             context.Context
		setupMock       func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError     bool
		expectNilResult bool
	}{
		{
			name: "happy path",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("parquet"), map[string]string{"meta": "data"}, nil)
				parquet.EXPECT().ReadStructsFromParquet(mock.Anything, []byte("parquet")).Return([]entity.TestCase{{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed}}, nil)
			},
			expectError:     false,
			expectNilResult: false,
		},
		{
			name:            "empty key",
			key:             "",
			ctx:             ctx,
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "nil context",
			key:             "valid key",
			ctx:             nil,
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "download error",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return(nil, nil, errors.New("download error"))
			},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "read parquet error",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("parquet"), map[string]string{}, nil)
				parquet.EXPECT().ReadStructsFromParquet(mock.Anything, []byte("parquet")).Return(nil, errors.New("read error"))
			},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "no data found",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("parquet"), map[string]string{}, nil)
				parquet.EXPECT().ReadStructsFromParquet(mock.Anything, []byte("parquet")).Return([]entity.TestCase{}, nil)
			},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "validation fails",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("parquet"), map[string]string{}, nil)
				parquet.EXPECT().ReadStructsFromParquet(mock.Anything, []byte("parquet")).Return([]entity.TestCase{{}}, nil)
			},
			expectError:     true,
			expectNilResult: true,
		},
	}
}

func TestReadAllMetadata(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	prefix := "testcase/"

	for _, test := range testcaseReadAllMetadataTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.setupMock != nil {
				test.setupMock(mockS3, mockParquet)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
				s3Prefix:       prefix,
			}

			data, err := repo.ReadAllMetadata(test.ctx)

			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedReturn, data)
			}
		})
	}
}

func testcaseReadAllMetadataTestCaseProvider() []struct {
	name           string
	ctx            context.Context
	setupMock      func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectedReturn []*entity.TestcaseMetadata
	expectError    bool
} {
	ctx := context.Background()
	return []struct {
		name           string
		ctx            context.Context
		setupMock      func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectedReturn []*entity.TestcaseMetadata
		expectError    bool
	}{
		{
			name:        "nil context",
			ctx:         nil,
			expectError: true,
		},
		{
			name: "listing of parquetfiles failed",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().ListParquetFiles(mock.Anything, "testcase/").Return(nil, errors.New("listing failed")).Once()
			},
			expectError: true,
		},
		{
			name: "download failed with single file",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().ListParquetFiles(mock.Anything, "testcase/").Return([]string{"key1"}, nil).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key1").Return(nil, nil, errors.New("download failed")).Once()
			},
			expectedReturn: []*entity.TestcaseMetadata{},
			expectError:    false,
		},
		{
			name: "download failed with multiple files - best effort",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().ListParquetFiles(mock.Anything, "testcase/").Return([]string{"key1", "key2", "key3"}, nil).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key1").Return([]byte("data1"), map[string]string{
					"author": "user1", "created": "123", "updated": "456", "name": "test1",
				}, nil).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key2").Return(nil, nil, errors.New("download failed")).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key3").Return([]byte("data3"), map[string]string{
					"author": "user3", "created": "789", "updated": "012", "name": "test3",
				}, nil).Once()
			},
			expectedReturn: []*entity.TestcaseMetadata{
				{Key: "key1", Author: "user1", Created: "123", Updated: "456", Name: "test1"},
				{Key: "key3", Author: "user3", Created: "789", Updated: "012", Name: "test3"},
			},
			expectError: false,
		},
		{
			name: "happy path with multiple files",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().ListParquetFiles(mock.Anything, "testcase/").Return([]string{"key1", "key2"}, nil).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key1").Return([]byte("data1"), map[string]string{
					"author": "user1", "created": "123", "updated": "456", "name": "test1",
				}, nil).Once()
				s3.EXPECT().DownloadParquetFile(mock.Anything, "key2").Return([]byte("data2"), map[string]string{
					"author": "user2", "created": "789", "updated": "012", "name": "test2",
				}, nil).Once()
			},
			expectedReturn: []*entity.TestcaseMetadata{
				{Key: "key1", Author: "user1", Created: "123", Updated: "456", Name: "test1"},
				{Key: "key2", Author: "user2", Created: "789", Updated: "012", Name: "test2"},
			},
			expectError: false,
		},
		{
			name: "happy path with empty list",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().ListParquetFiles(mock.Anything, "testcase/").Return([]string{}, nil).Once()
			},
			expectedReturn: []*entity.TestcaseMetadata{},
			expectError:    false,
		},
	}
}

// nolint:dupl
func TestUpdate(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range testcaseUpdateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.setupMock != nil {
				test.setupMock(mockS3, mockParquet)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
			}

			err := repo.Update(test.ctx, test.obj, test.key)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// nolint:funlen
func testcaseUpdateTestCaseProvider() []struct {
	name        string
	obj         *entity.TestCase
	key         string
	ctx         context.Context
	setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError bool
} {
	ctx := context.Background()
	return []struct {
		name        string
		obj         *entity.TestCase
		key         string
		ctx         context.Context
		setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError bool
	}{
		{
			name: "happy path",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("old data"), map[string]string{
					"testcase-id": "id",
					"author":      "original-author",
					"created":     "1234567890",
					"updated":     "1234567890",
					"name":        "test-name",
				}, nil)
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return([]byte("dummy parquet data"), nil)
				s3.EXPECT().UploadParquetFile(mock.Anything, "valid-key", []byte("dummy parquet data"), mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "nil obj",
			obj:         nil,
			key:         "valid-key",
			ctx:         ctx,
			expectError: true,
		},
		{
			name:        "nil context",
			obj:         &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:         "valid-key",
			ctx:         nil,
			expectError: true,
		},
		{
			name:        "validation fails",
			obj:         &entity.TestCase{},
			key:         "valid-key",
			ctx:         ctx,
			expectError: true,
		},
		{
			name:        "empty key",
			obj:         &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:         "",
			ctx:         ctx,
			expectError: true,
		},
		{
			name: "download metadata error",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return(nil, nil, errors.New("download error"))
			},
			expectError: true,
		},
		{
			name: "parquet write fails",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("old data"), map[string]string{
					"testcase-id": "id",
					"author":      "original-author",
					"created":     "1234567890",
					"updated":     "1234567890",
					"name":        "test-name",
				}, nil)
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return(nil, errors.New("parquet error"))
			},
			expectError: true,
		},
		{
			name: "s3 upload fails",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("old data"), map[string]string{
					"testcase-id": "id",
					"author":      "original-author",
					"created":     "1234567890",
					"updated":     "1234567890",
					"name":        "test-name",
				}, nil)
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return([]byte("dummy parquet data"), nil)
				s3.EXPECT().UploadParquetFile(mock.Anything, "valid-key", []byte("dummy parquet data"), mock.Anything).Return(errors.New("s3 error"))
			},
			expectError: true,
		},
	}
}

// nolint:dupl
func TestDeleteRemote(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		key         string
		ctx         context.Context
		setupMock   func(s3 *mocks.MockS3StorageWrapper)
		expectError bool
	}{
		{
			name: "happy path",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper) {
				s3.EXPECT().DeleteParquetFile(mock.Anything, "valid-key").Return(nil)
			},
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			ctx:         ctx,
			expectError: true,
		},
		{
			name:        "nil context",
			key:         "valid-key",
			ctx:         nil,
			expectError: true,
		},
		{
			name: "delete error",
			key:  "valid-key",
			ctx:  ctx,
			setupMock: func(s3 *mocks.MockS3StorageWrapper) {
				s3.EXPECT().DeleteParquetFile(mock.Anything, "valid-key").Return(errors.New("delete error"))
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.setupMock != nil {
				test.setupMock(mockS3)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
			}

			err := repo.Delete(test.ctx, test.key)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTestcaseData(t *testing.T) {
	tests := []struct {
		name        string
		obj         *entity.TestCase
		expectError bool
	}{
		{
			name: "valid TestCase",
			obj: &entity.TestCase{
				TestID:   "id",
				TestCode: entity.TestCode{Code: "code"},
				Status:   entity.TestStatusPassed,
			},
			expectError: false,
		},
		{
			name:        "nil obj",
			obj:         nil,
			expectError: true,
		},
		{
			name: "empty TestID",
			obj: &entity.TestCase{
				TestID:   "",
				TestCode: entity.TestCode{Code: "code"},
				Status:   entity.TestStatusPassed,
			},
			expectError: true,
		},
		{
			name: "empty TestCode.Code",
			obj: &entity.TestCase{
				TestID:   "id",
				TestCode: entity.TestCode{Code: ""},
				Status:   entity.TestStatusPassed,
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTestCaseData(test.obj)

			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateTestcaseKey(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)
	logger := slog.New(slog.DiscardHandler)
	prefix := "testcase/"
	testID := "test-id-123"

	repo := &testcaseStorageRepository{
		s3Wrapper:      mockS3,
		parquetWrapper: mockParquet,
		logger:         logger,
		s3Prefix:       prefix,
	}

	key := repo.generateTestCaseKey(testID)

	assert.NotEmpty(t, key, "key should not be empty")
	assert.True(t, strings.HasPrefix(key, prefix), "key should start with prefix '%s'", prefix)

	expectedStart := prefix + testID + "_"
	assert.True(t, strings.HasPrefix(key, expectedStart), "key should start with '%s'", expectedStart)
	assert.True(t, strings.HasSuffix(key, ".parquet"), "key should end with '.parquet'")

	rest := key[len(expectedStart):]
	rest = strings.TrimSuffix(rest, ".parquet")
	_, err := strconv.ParseInt(rest, 10, 64)
	assert.NoError(t, err, "key should contain a valid timestamp")
}

// nolint:dupl
func TestNewTestcaseStorageRepository(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.TestCase]
		prefix         string
		expectError    bool
	}{
		{
			name:           "all not nil",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			expectError:    false,
		},
		{
			name:           "nil logger",
			logger:         nil,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			expectError:    true,
		},
		{
			name:           "nil s3Wrapper",
			logger:         logger,
			s3Wrapper:      nil,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			expectError:    true,
		},
		{
			name:           "nil parquetWrapper",
			logger:         logger,
			s3Wrapper:      mockS3,
			prefix:         "prefix",
			parquetWrapper: nil,
			expectError:    true,
		},
		{
			name:           "empty prefix",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "",
			expectError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewTestcaseStorageRepository(test.logger, test.s3Wrapper, test.parquetWrapper, test.prefix, tracer)

			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}
