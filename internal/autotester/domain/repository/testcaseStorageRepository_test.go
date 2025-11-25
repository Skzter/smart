package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

// nolint:dupl
func TestCreateTestCaseStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range testcaseCreateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.TestCase]{}

			if test.obj != nil {
				mockParquet.On("WriteStructToParquet", *test.obj).
					Return(test.writeStructRet, test.writeStructErr)
			}
			if test.obj != nil && test.writeStructErr == nil {
				mockS3.On("UploadParquetFile", ctx, mock.Anything, test.writeStructRet, mock.Anything).
					Return(test.uploadRet)
			}

			repo := &testCaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
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
	name           string
	obj            *entity.TestCase
	userId         string
	writeStructRet []byte
	writeStructErr error
	uploadRet      error
	expectError    bool
} {
	return []struct {
		name           string
		obj            *entity.TestCase
		userId         string
		writeStructRet []byte
		writeStructErr error
		uploadRet      error
		expectError    bool
	}{
		{
			name:           "happy path",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:         "valid user",
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    false,
		},
		{
			name:        "nil obj",
			obj:         nil,
			expectError: true,
		},
		{
			name:           "validation fails",
			obj:            &entity.TestCase{},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name:           "userId validation fails",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:         "",
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name:           "parquet error",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:         "valid user",
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name:           "upload error",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			userId:         "valid user",
			writeStructRet: []byte("parquetdata"),
			uploadRet:      errors.New("upload error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestReadTestCaseStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range testcaseReadTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.TestCase]{}

			mockS3.On("DownloadParquetFile", ctx, test.key).
				Return(test.downloadRet, test.downloadMeta, test.downloadErr)

			mockParquet.On("ReadStructsFromParquet", test.downloadRet).
				Return(test.readStructsRet, test.readStructsErr)

			repo := &testCaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
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

func testcaseReadTestCaseProvider() []struct {
	name            string
	key             string
	downloadRet     []byte
	downloadMeta    map[string]string
	downloadErr     error
	readStructsRet  []entity.TestCase
	readStructsErr  error
	expectError     bool
	expectNilResult bool
} {
	return []struct {
		name            string
		key             string
		downloadRet     []byte
		downloadMeta    map[string]string
		downloadErr     error
		readStructsRet  []entity.TestCase
		readStructsErr  error
		expectError     bool
		expectNilResult bool
	}{
		{
			name:            "happy path",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.TestCase{{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed}},
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
			name:            "download error",
			key:             "valid-key",
			downloadErr:     errors.New("download error"),
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "read parquet error",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsErr:  errors.New("read error"),
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "no data found",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.TestCase{},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "validation fails",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.TestCase{{}},
			expectError:     true,
			expectNilResult: true,
		},
	}
}

// nolint:dupl
func TestUpdateTestCaseStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range testcaseUpdateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.TestCase]{}

			mockS3.On("FileExists", ctx, test.key).
				Return(test.fileExistsRet, test.fileExistsErr)

			if test.obj != nil && test.key != "" && test.fileExistsErr == nil && test.fileExistsRet {
				mockParquet.On("WriteStructToParquet", *test.obj).
					Return(test.writeStructRet, test.writeStructErr)
			}

			if test.obj != nil && test.key != "" && test.fileExistsErr == nil && test.fileExistsRet && test.writeStructErr == nil {
				mockS3.On("UploadParquetFile", ctx, test.key, test.writeStructRet, mock.Anything).
					Return(test.uploadRet)
			}

			repo := &testCaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
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
	name           string
	obj            *entity.TestCase
	key            string
	fileExistsRet  bool
	fileExistsErr  error
	writeStructRet []byte
	writeStructErr error
	uploadRet      error
	expectError    bool
} {
	return []struct {
		name           string
		obj            *entity.TestCase
		key            string
		fileExistsRet  bool
		fileExistsErr  error
		writeStructRet []byte
		writeStructErr error
		uploadRet      error
		expectError    bool
	}{
		{
			name:           "happy path",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructRet: []byte("dummy parquet data"),
			expectError:    false,
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
			name:          "file exists check error",
			obj:           &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:           "valid-key",
			fileExistsErr: errors.New("exists error"),
			expectError:   true,
		},
		{
			name:          "file does not exist",
			obj:           &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:           "valid-key",
			fileExistsRet: false,
			expectError:   true,
		},
		{
			name:           "parquet write fails",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name:           "s3 upload fails",
			obj:            &entity.TestCase{TestID: "id", TestCode: entity.TestCode{Code: "code"}, Status: entity.TestStatusPassed},
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructRet: []byte("dummy parquet data"),
			uploadRet:      errors.New("s3 error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestDeleteTestStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

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
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.TestCase]{}

			mockS3.On("DeleteParquetFile", ctx, test.key).
				Return(test.deleteRet)

			repo := &testCaseStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
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

func TestValidateTestCaseData(t *testing.T) {
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
			name: "empty Status",
			obj: &entity.TestCase{
				TestID:   "id",
				TestCode: entity.TestCode{Code: "code"},
				Status:   "",
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

func TestGenerateTestCaseKey(t *testing.T) {
	testID := "test-id-123"
	key := generateTestCaseKey(testID)

	if key == "" {
		t.Errorf("generateTestCaseKey() returned empty string")
	}
	const prefix = "autotester/testcase/"
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
func TestNewTestCaseStorageRepository(t *testing.T) {
	mockS3 := &mocks.MockS3StorageWrapper{}
	mockParquet := &mocks.MockParquetFileWrapper[entity.TestCase]{}
	logger := slog.Default()

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.TestCase]
		wantErr        bool
	}{
		{
			name:           "all not nil",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			wantErr:        false,
		},
		{
			name:           "nil logger",
			logger:         nil,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			wantErr:        true,
		},
		{
			name:           "nil s3Wrapper",
			logger:         logger,
			s3Wrapper:      nil,
			parquetWrapper: mockParquet,
			wantErr:        true,
		},
		{
			name:           "nil parquetWrapper",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: nil,
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewTestCaseStorageRepository(test.logger, test.s3Wrapper, test.parquetWrapper)
			if (err != nil) != test.wantErr {
				t.Errorf("NewTestCaseStorageRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewTestCaseStorageRepository() returned nil repository")
			}
		})
	}
}
