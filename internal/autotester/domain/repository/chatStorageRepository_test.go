package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

func validChat() *entity.Chat {
	return &entity.Chat{
		Id:            "chat123",
		UserId:        "user123",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastTest:      "test123",
		SystemPrompt:  "sys prompt",
		InitialPrompt: "usr prompt",
		Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
	}
}

// nolint:dupl
func TestCreateHistoryStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	for _, test := range historyCreateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}

			if test.obj != nil {
				mockParquet.On("WriteStructToParquet", *test.obj).
					Return(test.writeStructRet, test.writeStructErr)
			}
			if test.obj != nil && test.writeStructErr == nil {
				mockS3.On("UploadParquetFile", ctx, mock.Anything, test.writeStructRet, mock.Anything).
					Return(test.uploadRet)
			}

			repo := &chatStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
			}

			err := repo.Create(ctx, test.obj)
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

func historyCreateTestCaseProvider() []struct {
	name           string
	obj            *entity.Chat
	writeStructRet []byte
	writeStructErr error
	uploadRet      error
	expectError    bool
} {
	return []struct {
		name           string
		obj            *entity.Chat
		writeStructRet []byte
		writeStructErr error
		uploadRet      error
		expectError    bool
	}{
		{
			name:           "happy path",
			obj:            validChat(),
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
			name: "validation fails",
			obj: &entity.Chat{
				Id:            "",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			writeStructRet: []byte("parquetdata"),
			uploadRet:      nil,
			expectError:    true,
		},
		{
			name:           "parquet error",
			obj:            validChat(),
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name:           "upload error",
			obj:            validChat(),
			writeStructRet: []byte("parquetdata"),
			uploadRet:      errors.New("upload error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestReadHistoryStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	for _, test := range historyReadTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}

			mockS3.On("DownloadParquetFile", ctx, test.key).
				Return(test.downloadRet, test.downloadMeta, test.downloadErr)

			mockParquet.On("ReadStructsFromParquet", test.downloadRet).
				Return(test.readStructsRet, test.readStructsErr)

			repo := &chatStorageRepository{
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

func historyReadTestCaseProvider() []struct {
	name            string
	key             string
	downloadRet     []byte
	downloadMeta    map[string]string
	downloadErr     error
	readStructsRet  []entity.Chat
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
		readStructsRet  []entity.Chat
		readStructsErr  error
		expectError     bool
		expectNilResult bool
	}{
		{
			name:            "happy path",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.Chat{*validChat()},
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
			readStructsRet:  []entity.Chat{},
			expectError:     true,
			expectNilResult: true,
		},
		{
			name:            "validation fails",
			key:             "valid-key",
			downloadRet:     []byte("parquet"),
			downloadMeta:    map[string]string{},
			readStructsRet:  []entity.Chat{{}},
			expectError:     true,
			expectNilResult: true,
		},
	}
}

// nolint:dupl
func TestUpdateHistoryStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	for _, test := range historyUpdateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}

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

			repo := &chatStorageRepository{
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

func historyUpdateTestCaseProvider() []struct {
	name           string
	obj            *entity.Chat
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
		obj            *entity.Chat
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
			obj:            validChat(),
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
			obj:         &entity.Chat{},
			key:         "valid-key",
			expectError: true,
		},
		{
			name:        "empty key",
			obj:         validChat(),
			key:         "",
			expectError: true,
		},
		{
			name:          "file exists check error",
			obj:           validChat(),
			key:           "valid-key",
			fileExistsErr: errors.New("exists error"),
			expectError:   true,
		},
		{
			name:          "file does not exist",
			obj:           validChat(),
			key:           "valid-key",
			fileExistsRet: false,
			expectError:   true,
		},
		{
			name:           "parquet write fails",
			obj:            validChat(),
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructErr: errors.New("parquet error"),
			expectError:    true,
		},
		{
			name:           "s3 upload fails",
			obj:            validChat(),
			key:            "valid-key",
			fileExistsRet:  true,
			writeStructRet: []byte("dummy parquet data"),
			uploadRet:      errors.New("s3 error"),
			expectError:    true,
		},
	}
}

// nolint:dupl
func TestDeleteHistoryStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

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
			mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}

			mockS3.On("DeleteParquetFile", ctx, test.key).
				Return(test.deleteRet)

			repo := &chatStorageRepository{
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

// nolint:funlen
func TestValidateHistoryData(t *testing.T) {
	tests := []struct {
		name    string
		obj     *entity.Chat
		wantErr bool
	}{
		{
			name:    "valid Chat",
			obj:     validChat(),
			wantErr: false,
		},
		{
			name:    "nil obj",
			obj:     nil,
			wantErr: true,
		},
		{
			name: "empty id",
			obj: &entity.Chat{
				Id:            "",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty userId",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty initialPrompt",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty systemprompt",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "empty messages",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{},
			},
			wantErr: true,
		},
		{
			name: "message with empty body",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "user", Body: ""}},
			},
			wantErr: true,
		},
		{
			name: "message with empty id",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "", Role: "user", Body: "msg"}},
			},
			wantErr: true,
		},
		{
			name: "message with empty role",
			obj: &entity.Chat{
				Id:            "chat123",
				UserId:        "user123",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				LastTest:      "test123",
				SystemPrompt:  "sys prompt",
				InitialPrompt: "usr prompt",
				Messages:      []sharedEntity.Message{{Id: "id", Role: "", Body: "msg"}},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error
			if test.obj == nil {
				err = validateChat(nil)
			} else {
				err = validateChat(test.obj)
			}
			if (err != nil) != test.wantErr {
				t.Errorf("validateHistoryData() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestListAll(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	for _, test := range historyListAllTestCaseProvider(ctx) {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := &mocks.MockS3StorageWrapper{}
			mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}
			test.setupMocks(mockS3)

			repo := &chatStorageRepository{
				s3Wrapper:      mockS3,
				parquetWrapper: mockParquet,
				logger:         logger,
			}

			result, err := repo.ListAll(ctx)
			if (err != nil) != test.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(result) != test.wantCount {
				t.Errorf("ListAll() got %d entries, want %d", len(result), test.wantCount)
			}
		})
	}
}

// nolint:funlen
func historyListAllTestCaseProvider(ctx context.Context) []struct {
	name       string
	setupMocks func(
		mockS3 *mocks.MockS3StorageWrapper,
	)
	wantCount int
	wantErr   bool
} {
	now := fmt.Sprintf("%d", time.Now().UTC().Unix())
	metadata := map[string]string{
		"chat-id":       "id123",
		"user-id":       "userId123",
		"title":         "title123",
		"created-at":    now,
		"updated-at":    now,
		"message-count": "25",
	}
	return []struct {
		name       string
		setupMocks func(
			mockS3 *mocks.MockS3StorageWrapper,
		)
		wantCount int
		wantErr   bool
	}{
		{
			name: "all valid entries",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), metadata, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), metadata, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "list files error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return(nil, errors.New("list error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "no files found",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{}, nil)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name: "download error for one file",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return(nil, nil, errors.New("download error"))
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), metadata, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "file with incorrect created",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{
						"chat-id":       "id123",
						"user-id":       "userId123",
						"title":         "title123",
						"created-at":    "incorrect time",
						"updated-at":    now,
						"message-count": "25",
					}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), metadata, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "file with incorrect updated",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{
						"chat-id":       "id123",
						"user-id":       "userId123",
						"title":         "title123",
						"created-at":    now,
						"updated-at":    "incorrect time",
						"message-count": "25",
					}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), metadata, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "file with incorrect message count",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper) {
				mockS3.On("ListParquetFiles", ctx, "chat/").
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key1").
					Return([]byte("data1"), map[string]string{
						"chat-id":       "id123",
						"user-id":       "userId123",
						"title":         "title123",
						"created-at":    now,
						"updated-at":    now,
						"message-count": "incorrect number",
					}, nil)
				mockS3.On("DownloadParquetFile", ctx, "key2").
					Return([]byte("data2"), metadata, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}
}

// nolint:dupl
func TestGenerateChatKey(t *testing.T) {
	key := generateChatKey()

	if key == "" {
		t.Errorf("generateChatKey() returned empty string")
	}
	if len(key) < 20 {
		t.Errorf("generateChatKey() returned too short key: %s", key)
	}
	const prefix = "chat/"
	if len(key) < len(prefix) || key[:len(prefix)] != prefix {
		t.Errorf("generateChatKey() should start with '%s', got: %s", prefix, key)
	}
	if len(key) <= len(prefix) || key[len(prefix):] == "" {
		t.Errorf("generateChatKey() should contain a timestamp after prefix, got: %s", key)
	}
}

// nolint:dupl
func TestNewChatStorageRepository(t *testing.T) {
	mockS3 := &mocks.MockS3StorageWrapper{}
	mockParquet := &mocks.MockParquetFileWrapper[entity.Chat]{}
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.Chat]
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
			repo, err := NewChatStorageRepository(test.logger, test.s3Wrapper, test.parquetWrapper)
			if (err != nil) != test.wantErr {
				t.Errorf("NewChatStorageRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewChatStorageRepository() returned nil repository")
			}
		})
	}
}
