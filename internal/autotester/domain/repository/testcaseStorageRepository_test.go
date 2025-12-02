package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint:dupl
func TestCreate(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
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

			_, err := repo.Create(ctx, test.obj, test.userId)
			if test.expectError {
				if err == nil {
					t.Errorf("Create() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
			}
		})
	}
}

func testcaseCreateTestCaseProvider() []struct {
	name        string
	obj         *entity.TestCase
	userId      string
	setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError bool
} {
	return []struct {
		name        string
		obj         *entity.TestCase
		userId      string
		setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError bool
	}{
		{
			name:   "happy path",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
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
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "validation fails",
			obj:         &entity.TestCase{},
			userId:      "valid user",
			setupMock:   nil,
			expectError: true,
		},
		{
			name:        "userId validation fails",
			obj:         &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:      "",
			setupMock:   nil,
			expectError: true,
		},
		{
			name:   "parquet error",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return(nil, errors.New("parquet error"))
			},
			expectError: true,
		},
		{
			name:   "upload error",
			obj:    &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId: "valid user",
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
	ctx := context.Background()
	logger := slog.Default()
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

			result, err := repo.Read(ctx, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("Read() expected error but got none")
				}
				if !test.expectNilResult && result != nil {
					t.Errorf("Read() expected nil result on error, got: %+v", result)
				}
			} else {
				if err != nil {
					t.Errorf("Read() unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("Read() expected non-nil result on success")
				}
			}
		})
	}
}

func readRemoteTestcaseProvider() []struct {
	name            string
	key             string
	setupMock       func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError     bool
	expectNilResult bool
} {
	return []struct {
		name            string
		key             string
		setupMock       func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError     bool
		expectNilResult bool
	}{
		{
			name: "happy path",
			key:  "valid-key",
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
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "download error",
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return(nil, nil, errors.New("download error"))
			},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name: "read parquet error",
			key:  "valid-key",
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
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().DownloadParquetFile(mock.Anything, "valid-key").Return([]byte("parquet"), map[string]string{}, nil)
				parquet.EXPECT().ReadStructsFromParquet(mock.Anything, []byte("parquet")).Return([]entity.TestCase{{}}, nil)
			},
			expectError:     true,
			expectNilResult: true,
		},
	}
}

// nolint:dupl
func TestUpdate(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
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

			err := repo.Update(ctx, test.obj, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("Update() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Update() unexpected error: %v", err)
				}
			}
		})
	}
}

func testcaseUpdateTestCaseProvider() []struct {
	name        string
	obj         *entity.TestCase
	key         string
	setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
	expectError bool
} {
	return []struct {
		name        string
		obj         *entity.TestCase
		key         string
		setupMock   func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase])
		expectError bool
	}{
		{
			name: "happy path",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().FileExists(mock.Anything, "valid-key").Return(true, nil)
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return([]byte("dummy parquet data"), nil)
				s3.EXPECT().UploadParquetFile(mock.Anything, "valid-key", []byte("dummy parquet data"), mock.Anything).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "nil obj",
			obj:         nil,
			key:         "valid-key",
			expectError: true,
		},
		{
			name:        "validation fails",
			obj:         &entity.TestCase{},
			key:         "valid-key",
			expectError: true,
		},
		{
			name:        "empty key",
			obj:         &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:         "",
			expectError: true,
		},
		{
			name: "file exists check error",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().FileExists(mock.Anything, "valid-key").Return(false, errors.New("exists error"))
			},
			expectError: true,
		},
		{
			name: "file does not exist",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().FileExists(mock.Anything, "valid-key").Return(false, nil)
			},
			expectError: true,
		},
		{
			name: "parquet write fails",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().FileExists(mock.Anything, "valid-key").Return(true, nil)
				parquet.EXPECT().WriteStructToParquet(mock.Anything, mock.AnythingOfType("entity.TestCase")).Return(nil, errors.New("parquet error"))
			},
			expectError: true,
		},
		{
			name: "s3 upload fails",
			obj:  &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:  "valid-key",
			setupMock: func(s3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.TestCase]) {
				s3.EXPECT().FileExists(mock.Anything, "valid-key").Return(true, nil)
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
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		key         string
		deleteRet   error
		expectError bool
	}{
		{
			name:        "happy path",
			key:         "valid-key",
			deleteRet:   nil,
			expectError: false,
		},
		{
			name:        "empty key",
			key:         "",
			deleteRet:   nil,
			expectError: true,
		},
		{
			name:        "delete error",
			key:         "valid-key",
			deleteRet:   errors.New("delete error"),
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)

			if test.key != "" {
				mockS3.EXPECT().DeleteParquetFile(mock.Anything, test.key).
					Return(test.deleteRet)
			}

			repo := &testcaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
				tracer:         tracer,
			}

			err := repo.Delete(ctx, test.key)
			if test.expectError {
				if err == nil {
					t.Errorf("Delete() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Delete() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateTestcaseData(t *testing.T) {
	tests := []struct {
		name    string
		obj     *entity.TestCase
		wantErr bool
	}{
		{
			name: "valid TestCase",
			obj: &entity.TestCase{
				TestID:   "id",
				TestCode: entity.TestCode{Code: "code"},
				Status:   entity.TestStatusPassed,
			},
			wantErr: false,
		},
		{
			name:    "nil obj",
			obj:     nil,
			wantErr: true,
		},
		{
			name: "empty TestID",
			obj: &entity.TestCase{
				TestID:   "",
				TestCode: entity.TestCode{Code: "code"},
				Status:   entity.TestStatusPassed,
			},
			wantErr: true,
		},
		{
			name: "empty TestCode.Code",
			obj: &entity.TestCase{
				TestID:   "id",
				TestCode: entity.TestCode{Code: ""},
				Status:   entity.TestStatusPassed,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateTestCaseData(test.obj)
			if (err != nil) != test.wantErr {
				t.Errorf("validateTestCaseData() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestGenerateTestcaseKey(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)
	logger := slog.Default()
	prefix := "testcase/"
	testID := "test-id-123"
	repo := &testcaseStorageRepository{
		s3Wrapper:      mockS3,
		parquetWrapper: mockParquet,
		logger:         logger,
		s3Prefix:       prefix,
	}
	key := repo.generateTestCaseKey(testID)

	if key == "" {
		t.Errorf("generateTestCaseKey() returned empty string")
	}
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		t.Errorf("generateTestCaseKey() should start with '%s', got: %s", prefix, key)
	}

	expectedStart := prefix + testID + "_"
	if len(key) < len(expectedStart) || key[:len(expectedStart)] != expectedStart {
		t.Errorf("generateTestCaseKey() should start with '%s', got: %s", expectedStart, key)
	}

	if !strings.HasSuffix(key, ".parquet") {
		t.Errorf("generateTestCaseKey() should end with '.parquet', got: %s", key)
	}

	rest := key[len(expectedStart):]
	rest = strings.TrimSuffix(rest, ".parquet")
	if _, err := strconv.ParseInt(rest, 10, 64); err != nil {
		t.Errorf("generateTestCaseKey() should contain a valid timestamp, got: %s", rest)
	}
}

// nolint:dupl
func TestNewTestcaseStorageRepository(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.TestCase](t)
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.TestCase]
		prefix         string
		wantErr        bool
	}{
		{
			name:           "all not nil",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			wantErr:        false,
		},
		{
			name:           "nil logger",
			logger:         nil,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			wantErr:        true,
		},
		{
			name:           "nil s3Wrapper",
			logger:         logger,
			s3Wrapper:      nil,
			parquetWrapper: mockParquet,
			prefix:         "prefix",
			wantErr:        true,
		},
		{
			name:           "nil parquetWrapper",
			logger:         logger,
			s3Wrapper:      mockS3,
			prefix:         "prefix",
			parquetWrapper: nil,
			wantErr:        true,
		},
		{
			name:           "empty prefix",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			prefix:         "",
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewTestcaseStorageRepository(test.logger, test.s3Wrapper, test.parquetWrapper, test.prefix, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewTestCaseStorageRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewTestCaseStorageRepository() returned nil repository")
			}
		})
	}
}
