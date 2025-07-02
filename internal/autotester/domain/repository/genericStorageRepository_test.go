package repository

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

func TestCreate(t *testing.T) {
	testCreate[entity.TestCase](t)
	testCreate[entity.SessionSummary](t)
}

func testCreate[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname         string
		ctx              context.Context
		obj              *T
		validationFunc   func(*T) error
		mockParquetError error
		mockUploadError  error
		expectError      bool
	}{
		{
			testname:       fmt.Sprintf("%s: happy path", typeName),
			ctx:            context.Background(),
			obj:            func() *T { v := validTestObj[T](); return &v }(),
			validationFunc: func(obj *T) error { return nil },
			expectError:    false,
		},
		{
			testname:       fmt.Sprintf("%s: object validation fails", typeName),
			ctx:            context.Background(),
			obj:            func() *T { v := validTestObj[T](); return &v }(),
			validationFunc: func(obj *T) error { return fmt.Errorf("validation error") },
			expectError:    true,
		},
		{
			testname:    fmt.Sprintf("%s: nil obj", typeName),
			ctx:         context.Background(),
			obj:         nil,
			expectError: true,
		},
		{
			testname:         fmt.Sprintf("%s: parquet write fails", typeName),
			ctx:              context.Background(),
			obj:              func() *T { v := validTestObj[T](); return &v }(),
			validationFunc:   func(obj *T) error { return nil },
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        fmt.Sprintf("%s: s3 upload fails", typeName),
			ctx:             context.Background(),
			obj:             func() *T { v := validTestObj[T](); return &v }(),
			validationFunc:  func(obj *T) error { return nil },
			mockUploadError: fmt.Errorf("s3 error"),
			expectError:     true,
		},
	}

	for _, test := range tests {
		mockParquet := mocks.NewMockParquetFileWrapper[T](t)
		mockS3 := mocks.NewMockS3StorageWrapper(t)

		if test.obj != nil && test.validationFunc == nil {
			t.Fatalf("[%s] %s: validationFunc must not be nil", typeName, test.testname)
		}

		if test.obj != nil && test.validationFunc != nil && test.validationFunc(test.obj) == nil {
			mockParquet.On("WriteStructToParquet", mock.Anything).
				Return([]byte("dummy parquet data"), test.mockParquetError)

			if test.mockParquetError == nil {
				mockS3.On("UploadParquetFile", mock.Anything, mock.Anything, []byte("dummy parquet data"), mock.Anything).
					Return(test.mockUploadError)
			}
		}

		repo := &GenericStorageRepository[T]{
			logger:               testLogger(),
			parquetWrapper:       mockParquet,
			s3Wrapper:            mockS3,
			structValidationFunc: test.validationFunc,
		}

		t.Run(test.testname, func(t *testing.T) {
			key, err := repo.Create(test.ctx, test.obj)

			if test.expectError {
				require.Error(t, err, "[%s] %s: expected error but got none", typeName, test.testname)
				require.Equal(t, "", key, "[%s] %s: expected empty key on error, but got: %s", typeName, test.testname, key)
			} else {
				require.NoError(t, err, "[%s] %s: did not expect error but got: %v", typeName, test.testname, err)
				require.NotEmpty(t, key, "[%s] %s: expected non-empty key on success", typeName, test.testname)
			}

			mockParquet.AssertExpectations(t)
			mockS3.AssertExpectations(t)
		})
	}
}

func TestUpdate(t *testing.T) {
	testUpdate[entity.TestCase](t)
	testUpdate[entity.SessionSummary](t)
}

func testUpdate[T any](t *testing.T) {
	t.Helper()

	tests := getUpdateTests[T]()

	for _, test := range tests {
		mockParquet := mocks.NewMockParquetFileWrapper[T](t)
		mockS3 := mocks.NewMockS3StorageWrapper(t)

		mockParquet.
			On("WriteStructToParquet", mock.AnythingOfType(reflect.TypeOf(*new(T)).Name())).
			Return([]byte("dummy parquet data"), test.mockParquetError).
			Maybe()

		mockS3.
			On("FileExists", mock.Anything, test.key).
			Return(test.fileExists, test.mockFileExists).
			Maybe()

		if test.mockParquetError == nil && test.fileExists {
			mockS3.On("UploadParquetFile",
				mock.Anything,
				test.key,
				mock.AnythingOfType("[]uint8"),
				mock.MatchedBy(func(m map[string]string) bool { return true }),
			).Return(test.mockUploadError)
		}

		repo := &GenericStorageRepository[T]{
			logger:               testLogger(),
			parquetWrapper:       mockParquet,
			s3Wrapper:            mockS3,
			structValidationFunc: test.validationFunc,
		}

		t.Run(test.testname, func(t *testing.T) {
			err := repo.Update(test.ctx, test.obj, test.key)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockParquet.AssertExpectations(t)
			mockS3.AssertExpectations(t)
		})
	}
}

type updateTestCase[T any] struct {
	testname         string
	ctx              context.Context
	obj              *T
	validationFunc   func(*T) error
	key              string
	mockFileExists   error
	fileExists       bool
	mockParquetError error
	mockUploadError  error
	expectError      bool
}

func getUpdateTests[T any]() []updateTestCase[T] {
	typeName := reflect.TypeOf(*new(T)).Name()
	return []updateTestCase[T]{
		{
			testname:    fmt.Sprintf("%s: nil object returns error", typeName),
			ctx:         context.Background(),
			obj:         nil,
			key:         "valid-key",
			expectError: true,
		},
		{
			testname:       fmt.Sprintf("%s: object validation fails", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			validationFunc: func(obj *T) error { return fmt.Errorf("validation error") },
			key:            "valid-key",
			expectError:    true,
		},
		{
			testname:       fmt.Sprintf("%s: empty key returns error", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			validationFunc: func(obj *T) error { return nil },
			key:            "",
			expectError:    true,
		},
		{
			testname:       fmt.Sprintf("%s: FileExists returns error", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			validationFunc: func(obj *T) error { return nil },
			key:            "valid-key",
			mockFileExists: fmt.Errorf("file exists error"),
			expectError:    true,
		},
		{
			testname:       fmt.Sprintf("%s: File does not exist", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			validationFunc: func(obj *T) error { return nil },
			key:            "valid-key",
			fileExists:     false,
			expectError:    true,
		},
		{
			testname:         fmt.Sprintf("%s: Parquet serialization fails", typeName),
			ctx:              context.Background(),
			obj:              new(T),
			validationFunc:   func(obj *T) error { return nil },
			key:              "valid-key",
			fileExists:       true,
			mockParquetError: fmt.Errorf("parquet error"),
			expectError:      true,
		},
		{
			testname:        fmt.Sprintf("%s: s3 upload fails", typeName),
			ctx:             context.Background(),
			obj:             new(T),
			validationFunc:  func(obj *T) error { return nil },
			key:             "valid-key",
			fileExists:      true,
			mockUploadError: fmt.Errorf("upload error"),
			expectError:     true,
		},
		{
			testname:       fmt.Sprintf("%s: successful update", typeName),
			ctx:            context.Background(),
			obj:            new(T),
			validationFunc: func(obj *T) error { return nil },
			key:            "valid-key",
			fileExists:     true,
			expectError:    false,
		},
	}
}

func TestRead(t *testing.T) {
	testRead[entity.TestCase](t)
	testRead[entity.SessionSummary](t)
}

func testRead[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname             string
		ctx                  context.Context
		key                  string
		mockDownloadError    error
		mockReadParquetError error
		downloadedData       []byte
		readParquetData      []T
		validateFunc         func(*T) error
		expectError          bool
	}{
		{
			testname:    fmt.Sprintf("%s: empty key", typeName),
			ctx:         context.Background(),
			key:         "",
			expectError: true,
		},
		{
			testname:          fmt.Sprintf("%s: download fails", typeName),
			ctx:               context.Background(),
			key:               "test-key",
			mockDownloadError: fmt.Errorf("download failed"),
			expectError:       true,
		},
		{
			testname:             fmt.Sprintf("%s: read parquet fails", typeName),
			ctx:                  context.Background(),
			key:                  "test-key",
			downloadedData:       []byte("dummy"),
			mockReadParquetError: fmt.Errorf("read failed"),
			expectError:          true,
		},
		{
			testname:        fmt.Sprintf("%s: no data found", typeName),
			ctx:             context.Background(),
			key:             "test-key",
			downloadedData:  []byte("dummy"),
			readParquetData: []T{},
			expectError:     true,
		},
		{
			testname:        fmt.Sprintf("%s: read empty struct", typeName),
			ctx:             context.Background(),
			key:             "test-key",
			downloadedData:  []byte("dummy"),
			readParquetData: []T{*new(T)},
			validateFunc:    func(obj *T) error { return fmt.Errorf("validation error") },
			expectError:     true,
		},
		{
			testname:        fmt.Sprintf("%s: success", typeName),
			ctx:             context.Background(),
			key:             "test-key",
			downloadedData:  []byte("dummy"),
			readParquetData: []T{validTestObj[T]()},
			validateFunc:    func(obj *T) error { return nil },
			expectError:     false,
		},
	}

	for _, test := range tests {
		mockParquet := mocks.NewMockParquetFileWrapper[T](t)
		mockS3 := mocks.NewMockS3StorageWrapper(t)

		if test.key != "" {
			mockS3.On("DownloadParquetFile", mock.Anything, test.key).
				Return(test.downloadedData, map[string]string{}, test.mockDownloadError)
		}
		if test.downloadedData != nil && test.mockDownloadError == nil {
			mockParquet.On("ReadStructsFromParquet", test.downloadedData).
				Return(test.readParquetData, test.mockReadParquetError)
		}

		repo := &GenericStorageRepository[T]{
			logger:               testLogger(),
			parquetWrapper:       mockParquet,
			s3Wrapper:            mockS3,
			structValidationFunc: test.validateFunc,
		}

		t.Run(test.testname, func(t *testing.T) {
			obj, err := repo.Read(test.ctx, test.key)

			if test.expectError {
				require.Error(t, err, "[%s] %s: expected error but got none", typeName, test.testname)
			} else {
				require.NoError(t, err, "[%s] %s: did not expect error but got: %v", typeName, test.testname, err)
				require.NotNil(t, obj, "[%s] %s: expected non-nil object", typeName, test.testname)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testDelete[entity.TestCase](t)
	testDelete[entity.SessionSummary](t)
}

func testDelete[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname        string
		ctx             context.Context
		key             string
		mockDeleteError error
		expectError     bool
	}{
		{
			testname:    fmt.Sprintf("%s: empty key returns error", typeName),
			ctx:         context.Background(),
			key:         "",
			expectError: true,
		},
		{
			testname:        fmt.Sprintf("%s: s3 delete fails", typeName),
			ctx:             context.Background(),
			key:             "valid-key",
			mockDeleteError: fmt.Errorf("delete failed"),
			expectError:     true,
		},
		{
			testname:    fmt.Sprintf("%s: successful delete", typeName),
			ctx:         context.Background(),
			key:         "valid-key",
			expectError: false,
		},
	}

	for _, test := range tests {
		mockS3 := mocks.NewMockS3StorageWrapper(t)

		if test.key != "" {
			mockS3.
				On("DeleteParquetFile", mock.Anything, test.key).
				Return(test.mockDeleteError).
				Maybe()
		}

		repo := &GenericStorageRepository[T]{
			logger:    testLogger(),
			s3Wrapper: mockS3,
		}

		t.Run(test.testname, func(t *testing.T) {
			err := repo.Delete(test.ctx, test.key)

			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestGenerateKey(t *testing.T) {
	testGenerateKey[entity.TestCase](t)
	testGenerateKey[entity.SessionSummary](t)
}

func testGenerateKey[T any](t *testing.T) {
	t.Helper()

	typeName := reflect.TypeOf(*new(T)).Name()

	tests := []struct {
		testname string
		input    any
		wantType string
	}{
		{
			testname: fmt.Sprintf("%s: value", typeName),
			input:    *new(T),
			wantType: typeName,
		},
		{
			testname: fmt.Sprintf("%s: pointer", typeName),
			input:    new(T),
			wantType: typeName,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			key := generateKey(test.input)

			parts := strings.Split(key, "/")
			if len(parts) != 2 {
				t.Errorf("generateKey() = %q, expected format 'folder/filename'", key)
			}

			folder := parts[0]
			filename := parts[1]

			if folder != test.wantType {
				t.Errorf("folder = %q, want %q", folder, test.wantType)
			}

			expectedPrefix := fmt.Sprintf("%s_", test.wantType)
			expectedSuffix := ".parquet"

			if !strings.HasPrefix(filename, expectedPrefix) {
				t.Errorf("filename = %q, expected prefix %q", filename, expectedPrefix)
			}

			timestamp := strings.TrimSuffix(strings.TrimPrefix(filename, expectedPrefix), expectedSuffix)
			if len(timestamp) != 17 {
				t.Errorf("timestamp = %q, expected length 17", timestamp)
			}
		})
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func validTestObj[T any]() T {
	var zero T
	switch any(zero).(type) {
	case entity.TestCase:
		return any(entity.TestCase{
			TestID:      "id",
			Description: "desc",
			TestCode:    entity.TestCode{Code: "code"},
			Status:      entity.TestStatusPassed,
		}).(T)
	case entity.SessionSummary:
		return any(entity.SessionSummary{
			Summary:   "summary",
			CreatedAt: time.Now(),
			Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
		}).(T)
	default:
		panic("validTestObj: unsupported type")
	}
}
