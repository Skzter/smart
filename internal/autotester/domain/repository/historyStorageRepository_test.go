package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

// nolint:dupl
func TestCreate_historyStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range getCreateHistoryTestcases() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.SessionSummary]{}

			if test.obj != nil {
				mockParquet.On("WriteStructToParquet", *test.obj).
					Return(test.writeStructRet, test.writeStructErr)
			}
			if test.obj != nil && test.writeStructErr == nil {
				mockS3.On("UploadParquetFile", ctx, mock.Anything, test.writeStructRet, mock.Anything).
					Return(test.uploadRet)
			}

			repo := &sessionSummaryStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
			}

			key, err := repo.Create(ctx, test.obj)
			if test.expectError {
				if err == nil {
					t.Errorf("Create() expected error but got none")
				}
				if key != "" {
					t.Errorf("Create() expected empty key on error, got: %s", key)
				}
			} else {
				if err != nil {
					t.Errorf("Create() unexpected error: %v", err)
				}
				if key == "" {
					t.Errorf("Create() expected non-empty key on success")
				}
			}
		})
	}
}

func getCreateHistoryTestcases() []struct {
	name           string
	obj            *entity.SessionSummary
	writeStructRet []byte
	writeStructErr error
	uploadRet      error
	expectError    bool
} {
	return []struct {
		name           string
		obj            *entity.SessionSummary
		writeStructRet []byte
		writeStructErr error
		uploadRet      error
		expectError    bool
	}{
		{
			name: "happy path",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
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
			name: "validation fails (empty summary)",
			obj: &entity.SessionSummary{
				Summary:   "",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name: "validation fails (nil messages)",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  nil,
			},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name: "validation fails (messages contains nil)",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{nil},
			},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name: "parquet error",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name: "upload error",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      errors.New("upload error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestRead_historyStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range getReadHistoryTestcases() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.SessionSummary]{}

			mockS3.On("DownloadParquetFile", ctx, test.key).
				Return(test.downloadRet, test.downloadMeta, test.downloadErr)

			mockParquet.On("ReadStructsFromParquet", test.downloadRet).
				Return(test.readStructsRet, test.readStructsErr)

			repo := &sessionSummaryStorageRepository{
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

func getReadHistoryTestcases() []struct {
	name            string
	key             string
	downloadRet     []byte
	downloadMeta    map[string]string
	downloadErr     error
	readStructsRet  []entity.SessionSummary
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
		readStructsRet  []entity.SessionSummary
		readStructsErr  error
		expectError     bool
		expectNilResult bool
	}{
		{
			name:            "happy path",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.SessionSummary{{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}}},
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
			readStructsRet:  []entity.SessionSummary{},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "validation fails",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.SessionSummary{{}},
			expectError:     true,
			expectNilResult: true,
		},
	}
}

// nolint:dupl
func TestUpdate_historyStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, test := range getUpdateHistoryTestcases() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.SessionSummary]{}

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

			repo := &sessionSummaryStorageRepository{
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

func getUpdateHistoryTestcases() []struct {
	name           string
	obj            *entity.SessionSummary
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
		obj            *entity.SessionSummary
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
			obj:            &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
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
			obj:         &entity.SessionSummary{},
			key:         "valid-key",
			expectError: true,
		},
		{
			name:        "empty key",
			obj:         &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
			key:         "",
			expectError: true,
		},
		{
			name:          "file exists check error",
			obj:           &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
			key:           "valid-key",
			fileExistsErr: errors.New("exists error"),
			expectError:   true,
		},
		{
			name:          "file does not exist",
			obj:           &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
			key:           "valid-key",
			fileExistsRet: false,
			expectError:   true,
		},
		{
			name:           "parquet write fails",
			obj:            &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name:           "s3 upload fails",
			obj:            &entity.SessionSummary{Summary: "summary", CreatedAt: time.Now(), Messages: []*entity.Message{{Actor: "user", MessageBody: "msg"}}},
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructRet: []byte("dummy parquet data"),
			uploadRet:      errors.New("s3 error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestDelete_historyStorage(t *testing.T) {
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
			mockParquet := &mocks.MockParquetFileWrapper[entity.SessionSummary]{}

			mockS3.On("DeleteParquetFile", ctx, test.key).
				Return(test.deleteRet)

			repo := &sessionSummaryStorageRepository{
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

func Test_sessionSummaryValidationFunc(t *testing.T) {
	tests := []struct {
		name    string
		obj     *entity.SessionSummary
		wantErr bool
	}{
		{
			name: "valid SessionSummary",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			wantErr: false,
		},
		{
			name:    "nil obj",
			obj:     nil,
			wantErr: true,
		},
		{
			name: "empty Summary",
			obj: &entity.SessionSummary{
				Summary:   "",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "nil Messages",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  nil,
			},
			wantErr: true,
		},
		{
			name: "Messages contains nil",
			obj: &entity.SessionSummary{
				Summary:   "summary",
				CreatedAt: time.Now(),
				Messages:  []*entity.Message{nil},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.obj == nil {
				err = sessionSummaryValidationFunc(nil)
			} else {
				err = sessionSummaryValidationFunc(test.obj)
			}
			if (err != nil) != test.wantErr {
				t.Errorf("sessionSummaryValidationFunc() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func Test_ListAll(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	for _, tt := range getTestListAllCases(ctx) {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.SessionSummary]{}
			tt.setupMocks(mockS3, mockParquet)

			repo := &sessionSummaryStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
			}

			result := repo.ListAll(ctx)
			if len(result) != tt.wantCount {
				t.Errorf("ListAll() got %d entries, want %d", len(result), tt.wantCount)
			}
		})
	}
}

// nolint:funlen
func getTestListAllCases(ctx context.Context) []struct {
	name       string
	setupMocks func(
		mockS3 *mocks.MockS3StorageWrapper,
		mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary],
	)
	wantCount int
} {
	now := time.Now()
	validSummary := entity.SessionSummary{
		Summary:   "valid",
		CreatedAt: now,
		Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
	}
	invalidSummary := entity.SessionSummary{
		Summary:   "",
		CreatedAt: now,
		Messages:  []*entity.Message{{Actor: "user", MessageBody: "msg"}},
	}

	return []struct {
		name       string
		setupMocks func(
			mockS3 *mocks.MockS3StorageWrapper,
			mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary],
		)
		wantCount int
	}{
		{
			name: "all valid entries",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data1")).
					Return([]entity.SessionSummary{validSummary}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data2")).
					Return([]entity.SessionSummary{validSummary}, nil)
			},
			wantCount: 2,
		},
		{
			name: "list files error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return(nil, errors.New("list error"))
			},
			wantCount: 0,
		},
		{
			name: "no files found",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return([]string{}, nil)
			},
			wantCount: 0,
		},
		{
			name: "download error for one file",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return(nil, nil, errors.New("download error"))
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data2")).
					Return([]entity.SessionSummary{validSummary}, nil)
			},
			wantCount: 1,
		},
		{
			name: "parquet read error for one file",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data1")).
					Return(nil, errors.New("parquet error"))
				mockParquet.On("ReadStructsFromParquet", []byte("data2")).
					Return([]entity.SessionSummary{validSummary}, nil)
			},
			wantCount: 1,
		},
		{
			name: "invalid summary is skipped",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, mockParquet *mocks.MockParquetFileWrapper[entity.SessionSummary]) {
				mockS3.On("ListParquetFiles", ctx, "sessionSummary/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data1")).
					Return([]entity.SessionSummary{invalidSummary}, nil)
				mockParquet.On("ReadStructsFromParquet", []byte("data2")).
					Return([]entity.SessionSummary{validSummary}, nil)
			},
			wantCount: 1,
		},
	}
}

// nolint:dupl
func Test_generateSessionSummaryKey(t *testing.T) {
	key := generateSessionSummaryKey()

	if key == "" {
		t.Errorf("generateSessionSummaryKey() returned empty string")
	}
	if len(key) < 25 {
		t.Errorf("generateSessionSummaryKey() returned too short key: %s", key)
	}
	const prefix = "sessionSummary/"
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		t.Errorf("generateSessionSummaryKey() should start with '%s', got: %s", prefix, key)
	}
	if len(key) <= len(prefix) || key[len(prefix):] == "" {
		t.Errorf("generateSessionSummaryKey() should contain a timestamp after prefix, got: %s", key)
	}
}

func TestNewSessionSummaryStorageRepository(t *testing.T) {
	tests := []struct {
		name    string
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "default logger",
			logger:  slog.Default(),
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewSessionSummaryStorageRepository(tt.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSessionSummaryStorageRepository() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && repo == nil {
				t.Errorf("NewSessionSummaryStorageRepository() returned nil repository")
			}
		})
	}
}
