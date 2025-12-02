package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint:dupl
func TestCreateChatStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range chatCreateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Chat](t)
			mockSummaryParquet := mocks.NewMockParquetFileWrapper[entity.ChatSummary](t)

			if test.obj != nil {
				test.obj.Filter(entity.MessageTypeUser)
			}

			if test.chatWriteRet != nil {
				mockParquet.On("WriteStructToParquet", mock.Anything, *test.obj).Return(test.chatWriteRet...)
			}
			if test.summaryWriteRet != nil {
				mockSummaryParquet.On("WriteStructToParquet", mock.Anything, mock.Anything).Return(test.summaryWriteRet...)
			}
			if test.uploadRet != nil {
				mockS3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.uploadRet...)
			}

			repo, _ := NewChatStorageRepository(logger, mockS3, mockParquet, mockSummaryParquet, tracer)

			err := repo.Create(test.ctx, test.obj)
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

func chatCreateTestCaseProvider() []struct {
	name            string
	obj             *entity.Chat
	chatWriteRet    []any
	summaryWriteRet []any
	uploadRet       []any
	expectError     bool
	ctx             context.Context
} {
	return []struct {
		name            string
		obj             *entity.Chat
		chatWriteRet    []any
		summaryWriteRet []any
		uploadRet       []any
		expectError     bool
		ctx             context.Context
	}{
		{
			name:            "happy path",
			obj:             &entity.Chat{},
			chatWriteRet:    []any{[]byte("chat parqeut data"), nil},
			summaryWriteRet: []any{[]byte("summary parqeut data"), nil},
			uploadRet:       []any{nil},
			expectError:     false,
			ctx:             context.Background(),
		},
		{
			name:        "nil obj",
			ctx:         context.Background(),
			obj:         nil,
			expectError: true,
		},
		{
			name:        "nil ctx",
			obj:         &entity.Chat{},
			ctx:         nil,
			expectError: true,
		},
		{
			name:         "chat parquet error",
			obj:          &entity.Chat{},
			chatWriteRet: []any{nil, errors.New("err")},
			expectError:  true,
			ctx:          context.Background(),
		},
		{
			name:            "summary parquet error",
			obj:             &entity.Chat{},
			chatWriteRet:    []any{[]byte("chat parqeut data"), nil},
			summaryWriteRet: []any{nil, errors.New("err")},
			expectError:     true,
			ctx:             context.Background(),
		},
		{
			name:            "upload error",
			obj:             &entity.Chat{},
			chatWriteRet:    []any{[]byte("chat parqeut data"), nil},
			summaryWriteRet: []any{[]byte("summary parqeut data"), nil},
			uploadRet:       []any{errors.New("err")},
			expectError:     true,
			ctx:             context.Background(),
		},
	}
}

// nolint:dupl
func TestReadChatStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range chatReadTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Chat](t)
			mockSummaryParquet := mocks.NewMockParquetFileWrapper[entity.ChatSummary](t)

			key, _ := generateKeys("user", "chat")

			if test.downloadRet != nil {
				mockS3.On("DownloadParquetFile", mock.Anything, key).Return(test.downloadRet...)
			}

			if test.readStructsRet != nil {
				mockParquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).Return(test.readStructsRet...)
			}

			repo, _ := NewChatStorageRepository(logger, mockS3, mockParquet, mockSummaryParquet, tracer)

			result, err := repo.Read(test.ctx, "user", "chat")
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func chatReadTestCaseProvider() []struct {
	name           string
	downloadRet    []any
	readStructsRet []any
	expectError    bool
	ctx            context.Context
} {
	return []struct {
		name           string
		downloadRet    []any
		readStructsRet []any
		expectError    bool
		ctx            context.Context
	}{
		{
			name:           "happy path",
			downloadRet:    []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{[]entity.Chat{{}}, nil},
			expectError:    false,
			ctx:            context.Background(),
		},
		{
			name:        "nil ctx",
			expectError: true,
			ctx:         nil,
		},
		{
			name:        "download error",
			downloadRet: []any{nil, nil, errors.New("download error")},
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:           "read parquet error",
			downloadRet:    []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{nil, errors.New("read error")},
			expectError:    true,
			ctx:            context.Background(),
		},
		{
			name:           "no data found",
			downloadRet:    []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{[]entity.Chat{}, nil},
			expectError:    true,
			ctx:            context.Background(),
		},
	}
}

// nolint:dupl
func TestDeleteChatStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name             string
		deleteChatRet    []any
		deleteSummaryRet []any
		expectError      bool
		ctx              context.Context
	}{
		{
			name:             "happy path",
			deleteChatRet:    []any{nil},
			deleteSummaryRet: []any{nil},
			expectError:      false,
			ctx:              context.Background(),
		},
		{
			name:        "nil ctx",
			expectError: true,
			ctx:         nil,
		},
		{
			name:          "delete chat error",
			deleteChatRet: []any{errors.New("delete error")},
			expectError:   true,
			ctx:           context.Background(),
		},
		{
			name:             "delete chat error",
			deleteChatRet:    []any{nil},
			deleteSummaryRet: []any{errors.New("delete error")},
			expectError:      true,
			ctx:              context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Chat](t)
			mockSummaryParquet := mocks.NewMockParquetFileWrapper[entity.ChatSummary](t)
			chatkey, summarykey := generateKeys("user", "chat")
			if test.deleteChatRet != nil {
				mockS3.On("DeleteParquetFile", mock.Anything, chatkey).
					Return(test.deleteChatRet...)
			}
			if test.deleteSummaryRet != nil {
				mockS3.On("DeleteParquetFile", mock.Anything, summarykey).
					Return(test.deleteSummaryRet...)
			}

			repo, _ := NewChatStorageRepository(logger, mockS3, mockParquet, mockSummaryParquet, tracer)

			err := repo.Delete(test.ctx, "user", "chat")
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

func TestFindByUserID(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range findByUserIDTestCaseProvider(ctx) {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockSummaryParquet := mocks.NewMockParquetFileWrapper[entity.ChatSummary](t)

			test.setupMocks(mockS3, mockSummaryParquet)

			repo, _ := NewChatStorageRepository(logger, mockS3, mocks.NewMockParquetFileWrapper[entity.Chat](t), mockSummaryParquet, tracer)

			result, err := repo.FindByUserID(test.ctx, "user")
			if (err != nil) != test.wantErr {
				t.Errorf("ListAll() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(result) != test.wantCount {
				t.Errorf("FindByUserID() got %d entries, want %d", len(result), test.wantCount)
			}
		})
	}
}

// nolint:funlen
func findByUserIDTestCaseProvider(ctx context.Context) []struct {
	name       string
	setupMocks func(
		mockS3 *mocks.MockS3StorageWrapper,
		parquet *mocks.MockParquetFileWrapper[entity.ChatSummary],
	)
	wantCount int
	wantErr   bool
	ctx       context.Context
} {
	return []struct {
		name       string
		setupMocks func(
			mockS3 *mocks.MockS3StorageWrapper,
			parquet *mocks.MockParquetFileWrapper[entity.ChatSummary],
		)
		wantCount int
		wantErr   bool
		ctx       context.Context
	}{
		{
			name: "happy path",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", mock.Anything, "key1").
					Return([]byte("data1"), map[string]string{}, nil)
				mockS3.On("DownloadParquetFile", mock.Anything, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				parquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).
					Return([]entity.ChatSummary{{}}, nil)
			},
			wantCount: 2,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name:       "nil ctx",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {},
			wantCount:  0,
			wantErr:    true,
			ctx:        nil,
		},
		{
			name: "s3 list error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return(nil, errors.New("err"))
			},
			wantCount: 0,
			wantErr:   true,
			ctx:       ctx,
		},
		{
			name: "no keys found",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return([]string{}, nil)
			},
			wantCount: 0,
			wantErr:   true,
			ctx:       ctx,
		},
		{
			name: "s3 download error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", mock.Anything, "key1").
					Return(nil, nil, errors.New("err"))
				mockS3.On("DownloadParquetFile", mock.Anything, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				parquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).
					Return([]entity.ChatSummary{{}}, nil)
			},
			wantCount: 1,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name: "parquet error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", mock.Anything, "key1").
					Return(nil, nil, errors.New("err"))
				mockS3.On("DownloadParquetFile", mock.Anything, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				parquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).
					Return(nil, errors.New("err"))
			},
			wantCount: 0,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name: "incorrect number of structs in parquet file",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.ChatSummary]) {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).
					Return([]string{"key1", "key2"}, nil)
				mockS3.On("DownloadParquetFile", mock.Anything, "key1").
					Return(nil, nil, errors.New("err"))
				mockS3.On("DownloadParquetFile", mock.Anything, "key2").
					Return([]byte("data2"), map[string]string{}, nil)
				parquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).
					Return([]entity.ChatSummary{{}, {}}, nil)
			},
			wantCount: 0,
			wantErr:   false,
			ctx:       ctx,
		},
	}
}

// (generateChatKey removed in implementation)

// nolint:dupl
func TestNewChatStorageRepository(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockChatParquet := mocks.NewMockParquetFileWrapper[entity.Chat](t)
	mockSummaryParquet := mocks.NewMockParquetFileWrapper[entity.ChatSummary](t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name                  string
		logger                *slog.Logger
		s3Wrapper             service.S3StorageWrapper
		chatParquetWrapper    service.ParquetFileWrapper[entity.Chat]
		summaryParquetWrapper service.ParquetFileWrapper[entity.ChatSummary]
		wantErr               bool
	}{
		{
			name:                  "all not nil",
			logger:                logger,
			s3Wrapper:             mockS3,
			chatParquetWrapper:    mockChatParquet,
			summaryParquetWrapper: mockSummaryParquet,
			wantErr:               false,
		},
		{
			name:                  "nil logger",
			logger:                nil,
			s3Wrapper:             mockS3,
			chatParquetWrapper:    mockChatParquet,
			summaryParquetWrapper: mockSummaryParquet,
			wantErr:               true,
		},
		{
			name:                  "nil s3Wrapper",
			logger:                logger,
			s3Wrapper:             nil,
			chatParquetWrapper:    mockChatParquet,
			summaryParquetWrapper: mockSummaryParquet,
			wantErr:               true,
		},
		{
			name:                  "nil parquetWrapper",
			logger:                logger,
			s3Wrapper:             mockS3,
			chatParquetWrapper:    nil,
			summaryParquetWrapper: mockSummaryParquet,
			wantErr:               true,
		},
		{
			name:                  "nil summaryParquetWrapper",
			logger:                logger,
			s3Wrapper:             mockS3,
			chatParquetWrapper:    mockChatParquet,
			summaryParquetWrapper: nil,
			wantErr:               true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewChatStorageRepository(test.logger, test.s3Wrapper, test.chatParquetWrapper, test.summaryParquetWrapper, tracer)
			if (err != nil) != test.wantErr {
				t.Errorf("NewChatStorageRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewChatStorageRepository() returned nil repository")
			}
		})
	}
}
