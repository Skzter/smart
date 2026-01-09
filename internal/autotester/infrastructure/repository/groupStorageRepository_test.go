package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
)

// nolint:dupl
func TestCreateGroupStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range groupCreateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)

			if test.writeRet != nil {
				mockParquet.On("WriteStructToParquet", mock.Anything, *test.obj).Return(test.writeRet...)
			}
			if test.uploadRet != nil {
				mockS3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.uploadRet...)
			}

			repo, _ := NewGroupStorageRepository(logger, mockS3, mockParquet, tracer)

			err := repo.Create(test.ctx, test.obj)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockParquet.AssertExpectations(t)
			mockS3.AssertExpectations(t)
		})
	}
}

func groupCreateTestCaseProvider() []struct {
	name        string
	obj         *entity.Group
	writeRet    []any
	uploadRet   []any
	expectError bool
	ctx         context.Context
} {
	return []struct {
		name        string
		obj         *entity.Group
		writeRet    []any
		uploadRet   []any
		expectError bool
		ctx         context.Context
	}{
		{
			name: "happy path",
			obj: &entity.Group{
				Id:        "group-1",
				Name:      "Test Group",
				CreatedBy: "user-1",
			},
			writeRet:    []any{[]byte("parquet data"), nil},
			uploadRet:   []any{nil},
			expectError: false,
			ctx:         context.Background(),
		},
		{
			name:        "nil obj",
			ctx:         context.Background(),
			obj:         nil,
			expectError: true,
		},
		{
			name: "nil ctx",
			obj: &entity.Group{
				Id: "group-1",
			},
			ctx:         nil,
			expectError: true,
		},
		{
			name: "parquet write error",
			obj: &entity.Group{
				Id: "group-1",
			},
			writeRet:    []any{nil, errors.New("parquet write error")},
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name: "upload error",
			obj: &entity.Group{
				Id: "group-1",
			},
			writeRet:    []any{[]byte("parquet data"), nil},
			uploadRet:   []any{errors.New("upload error")},
			expectError: true,
			ctx:         context.Background(),
		},
	}
}

// nolint:dupl
func TestReadGroupStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range groupReadTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)

			if test.groupId != "" {
				key := generateKey(test.groupId)

				if test.downloadRet != nil {
					mockS3.On("DownloadParquetFile", mock.Anything, key).Return(test.downloadRet...)
				}

				if test.readStructsRet != nil {
					mockParquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).Return(test.readStructsRet...)
				}
			}

			repo, _ := NewGroupStorageRepository(logger, mockS3, mockParquet, tracer)

			result, err := repo.Read(test.ctx, test.groupId)
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func groupReadTestCaseProvider() []struct {
	name           string
	groupId        string
	downloadRet    []any
	readStructsRet []any
	expectError    bool
	ctx            context.Context
} {
	return []struct {
		name           string
		groupId        string
		downloadRet    []any
		readStructsRet []any
		expectError    bool
		ctx            context.Context
	}{
		{
			name:        "happy path",
			groupId:     "group-1",
			downloadRet: []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{[]entity.Group{{
				Id:   "group-1",
				Name: "Test Group",
			}}, nil},
			expectError: false,
			ctx:         context.Background(),
		},
		{
			name:        "nil ctx",
			groupId:     "group-1",
			expectError: true,
			ctx:         nil,
		},
		{
			name:        "empty groupId",
			groupId:     "",
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:        "download error",
			groupId:     "group-1",
			downloadRet: []any{nil, nil, errors.New("download error")},
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:           "read parquet error",
			groupId:        "group-1",
			downloadRet:    []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{nil, errors.New("read error")},
			expectError:    true,
			ctx:            context.Background(),
		},
		{
			name:           "no groups found",
			groupId:        "group-1",
			downloadRet:    []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{[]entity.Group{}, nil},
			expectError:    true,
			ctx:            context.Background(),
		},
		{
			name:        "multiple groups found",
			groupId:     "group-1",
			downloadRet: []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet: []any{[]entity.Group{
				{Id: "group-1"},
				{Id: "group-2"},
			}, nil},
			expectError: true,
			ctx:         context.Background(),
		},
	}
}

// nolint:dupl
func TestUpdateGroupStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range groupUpdateTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)

			if test.obj != nil {
				key := generateKey(test.obj.Id)
				// Read call for existence check
				if test.readDownloadRet != nil {
					mockS3.On("DownloadParquetFile", mock.Anything, key).Return(test.readDownloadRet...)
				}
				if test.readStructsRet != nil {
					mockParquet.On("ReadStructsFromParquet", mock.Anything, mock.Anything).Return(test.readStructsRet...)
				}
				// Write calls for update
				if test.writeRet != nil {
					mockParquet.On("WriteStructToParquet", mock.Anything, *test.obj).Return(test.writeRet...)
				}
				if test.uploadRet != nil {
					mockS3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.uploadRet...)
				}
			}

			repo, _ := NewGroupStorageRepository(logger, mockS3, mockParquet, tracer)

			err := repo.Update(test.ctx, test.obj)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func groupUpdateTestCaseProvider() []struct {
	name            string
	obj             *entity.Group
	readDownloadRet []any
	readStructsRet  []any
	writeRet        []any
	uploadRet       []any
	expectError     bool
	ctx             context.Context
} {
	return []struct {
		name            string
		obj             *entity.Group
		readDownloadRet []any
		readStructsRet  []any
		writeRet        []any
		uploadRet       []any
		expectError     bool
		ctx             context.Context
	}{
		{
			name: "happy path",
			obj: &entity.Group{
				Id:   "group-1",
				Name: "Updated Group",
			},
			readDownloadRet: []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet:  []any{[]entity.Group{{Id: "group-1"}}, nil},
			writeRet:        []any{[]byte("parquet data"), nil},
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
			name: "nil ctx",
			obj: &entity.Group{
				Id: "group-1",
			},
			ctx:         nil,
			expectError: true,
		},
		{
			name: "group not found",
			obj: &entity.Group{
				Id: "group-1",
			},
			readDownloadRet: []any{nil, nil, errors.New("not found")},
			expectError:     true,
			ctx:             context.Background(),
		},
		{
			name: "parquet write error",
			obj: &entity.Group{
				Id: "group-1",
			},
			readDownloadRet: []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet:  []any{[]entity.Group{{Id: "group-1"}}, nil},
			writeRet:        []any{nil, errors.New("parquet write error")},
			expectError:     true,
			ctx:             context.Background(),
		},
		{
			name: "upload error",
			obj: &entity.Group{
				Id: "group-1",
			},
			readDownloadRet: []any{[]byte("parquet"), map[string]string{}, nil},
			readStructsRet:  []any{[]entity.Group{{Id: "group-1"}}, nil},
			writeRet:        []any{[]byte("parquet data"), nil},
			uploadRet:       []any{errors.New("upload error")},
			expectError:     true,
			ctx:             context.Background(),
		},
	}
}

// nolint:dupl
func TestDeleteGroupStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		groupId     string
		deleteRet   []any
		expectError bool
		ctx         context.Context
	}{
		{
			name:        "happy path",
			groupId:     "group-1",
			deleteRet:   []any{nil},
			expectError: false,
			ctx:         context.Background(),
		},
		{
			name:        "nil ctx",
			groupId:     "group-1",
			expectError: true,
			ctx:         nil,
		},
		{
			name:        "empty groupId",
			groupId:     "",
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:        "delete error",
			groupId:     "group-1",
			deleteRet:   []any{errors.New("delete error")},
			expectError: true,
			ctx:         context.Background(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)

			if test.groupId != "" && test.ctx != nil {
				key := generateKey(test.groupId)
				if test.deleteRet != nil {
					mockS3.On("DeleteParquetFile", mock.Anything, key).Return(test.deleteRet...)
				}
			}

			repo, _ := NewGroupStorageRepository(logger, mockS3, mockParquet, tracer)

			err := repo.Delete(test.ctx, test.groupId)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func TestListAllGroupStorage(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	for _, test := range listAllGroupTestCaseProvider(ctx) {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)

			test.setupMocks(mockS3, mockParquet)

			repo, _ := NewGroupStorageRepository(logger, mockS3, mockParquet, tracer)

			result, err := repo.ListAll(test.ctx)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.wantCount, len(result))

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

// nolint:funlen
func listAllGroupTestCaseProvider(ctx context.Context) []struct {
	name       string
	setupMocks func(
		mockS3 *mocks.MockS3StorageWrapper,
		parquet *mocks.MockParquetFileWrapper[entity.Group],
	)
	wantCount int
	wantErr   bool
	ctx       context.Context
} {
	return []struct {
		name       string
		setupMocks func(
			mockS3 *mocks.MockS3StorageWrapper,
			parquet *mocks.MockParquetFileWrapper[entity.Group],
		)
		wantCount int
		wantErr   bool
		ctx       context.Context
	}{
		{
			name: "happy path",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return([]string{"group/group-1.parquet", "group/group-2.parquet"}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-1.parquet").
					Return([]byte("data1"), map[string]string{}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-2.parquet").
					Return([]byte("data2"), map[string]string{}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data1")).
					Return([]entity.Group{{Id: "group-1"}}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data2")).
					Return([]entity.Group{{Id: "group-2"}}, nil).Once()
			},
			wantCount: 2,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name:       "nil ctx",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {},
			wantCount:  0,
			wantErr:    true,
			ctx:        nil,
		},
		{
			name: "s3 list error",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return(nil, errors.New("list error")).Once()
			},
			wantCount: 0,
			wantErr:   true,
			ctx:       ctx,
		},
		{
			name: "no keys found",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return([]string{}, nil).Once()
			},
			wantCount: 0,
			wantErr:   true,
			ctx:       ctx,
		},
		{
			name: "s3 download error - continues with other files",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return([]string{"group/group-1.parquet", "group/group-2.parquet"}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-1.parquet").
					Return(nil, nil, errors.New("download error")).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-2.parquet").
					Return([]byte("data2"), map[string]string{}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data2")).
					Return([]entity.Group{{Id: "group-2"}}, nil).Once()
			},
			wantCount: 1,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name: "parquet error - continues with other files",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return([]string{"group/group-1.parquet", "group/group-2.parquet"}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-1.parquet").
					Return([]byte("data1"), map[string]string{}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-2.parquet").
					Return([]byte("data2"), map[string]string{}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data1")).
					Return(nil, errors.New("parquet error")).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data2")).
					Return([]entity.Group{{Id: "group-2"}}, nil).Once()
			},
			wantCount: 1,
			wantErr:   false,
			ctx:       ctx,
		},
		{
			name: "incorrect number of structs in parquet file - continues with other files",
			setupMocks: func(mockS3 *mocks.MockS3StorageWrapper, parquet *mocks.MockParquetFileWrapper[entity.Group]) {
				mockS3.On("ListParquetFiles", mock.Anything, prefixGroup).
					Return([]string{"group/group-1.parquet", "group/group-2.parquet"}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-1.parquet").
					Return([]byte("data1"), map[string]string{}, nil).Once()
				mockS3.On("DownloadParquetFile", mock.Anything, "group/group-2.parquet").
					Return([]byte("data2"), map[string]string{}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data1")).
					Return([]entity.Group{{Id: "group-1"}, {Id: "group-3"}}, nil).Once()
				parquet.On("ReadStructsFromParquet", mock.Anything, []byte("data2")).
					Return([]entity.Group{{Id: "group-2"}}, nil).Once()
			},
			wantCount: 1,
			wantErr:   false,
			ctx:       ctx,
		},
	}
}

// nolint:dupl
func TestNewGroupStorageRepository(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.Group](t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.Group]
		tracer         trace.Tracer
		wantErr        bool
	}{
		{
			name:           "all not nil",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			tracer:         tracer,
			wantErr:        false,
		},
		{
			name:           "nil logger",
			logger:         nil,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			tracer:         tracer,
			wantErr:        true,
		},
		{
			name:           "nil s3Wrapper",
			logger:         logger,
			s3Wrapper:      nil,
			parquetWrapper: mockParquet,
			tracer:         tracer,
			wantErr:        true,
		},
		{
			name:           "nil parquetWrapper",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: nil,
			tracer:         tracer,
			wantErr:        true,
		},
		{
			name:           "nil tracer",
			logger:         logger,
			s3Wrapper:      mockS3,
			parquetWrapper: mockParquet,
			tracer:         nil,
			wantErr:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewGroupStorageRepository(test.logger, test.s3Wrapper, test.parquetWrapper, test.tracer)
			if test.wantErr {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		groupId  string
		expected string
	}{
		{
			name:     "simple groupId",
			groupId:  "group-1",
			expected: "group/group-1.parquet",
		},
		{
			name:     "groupId with uuid",
			groupId:  "550e8400-e29b-41d4-a716-446655440000",
			expected: "group/550e8400-e29b-41d4-a716-446655440000.parquet",
		},
		{
			name:     "empty groupId",
			groupId:  "",
			expected: "group/.parquet",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := generateKey(test.groupId)
			assert.Equal(t, test.expected, result)
		})
	}
}
