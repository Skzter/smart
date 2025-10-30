package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
)

func uploadParquetFile(err error) *mockCall { return call("UploadParquetFile", 4, err) }
func writeStructToParquet(b []byte, err error) *mockCall {
	return call("WriteStructToParquet", 1, b, err)
}
func downloadParquetFile(b []byte, m map[string]string, err error) *mockCall {
	return call("DownloadParquetFile", 2, b, m, err)
}
func readStructsFromParquet(tle []entity.TagListEntity, err error) *mockCall {
	return call("ReadStructsFromParquet", 1, tle, err)
}
func fileExists(b bool, err error) *mockCall { return call("FileExists", 2, b, err) }

func call(name string, params int, returns ...any) *mockCall {
	return &mockCall{Name: name, Params: params, Returns: returns}
}

type mockCall struct {
	Name    string
	Params  int
	Returns []any
}

func ctx(t *testing.T) context.Context { return t.Context() }

func setupMocks(s3calls ...*mockCall) func(...*mockCall) func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
	return func(parquetCalls ...*mockCall) func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
		return func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
			s3 := mocks.NewMockS3StorageWrapper(t)
			parquet := mocks.NewMockParquetFileWrapper[entity.TagListEntity](t)
			for _, arg := range s3calls {
				args := make([]any, arg.Params)
				for i := range args {
					args[i] = mock.Anything
				}
				s3.On(arg.Name, args...).Return(arg.Returns...)
			}
			for _, arg := range parquetCalls {
				args := make([]any, arg.Params)
				for i := range args {
					args[i] = mock.Anything
				}
				parquet.On(arg.Name, args...).Return(arg.Returns...)
			}
			return s3, parquet
		}
	}
}

func TestCreateTaglist(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		taglist      *entity.TagListEntity
		ctx          func(t *testing.T) context.Context
		expectsError bool
	}{
		{
			name:         "happy path",
			mockSetup:    setupMocks(uploadParquetFile(nil))(writeStructToParquet([]byte{}, nil)),
			taglist:      &entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: false,
			ctx:          ctx,
		},
		{
			name:         "nil ctx",
			mockSetup:    setupMocks()(),
			taglist:      &entity.TagListEntity{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "validation error",
			mockSetup:    setupMocks()(),
			taglist:      &entity.TagListEntity{},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "parquet error",
			mockSetup:    setupMocks()(writeStructToParquet(nil, errors.New("err"))),
			taglist:      &entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "s3 error",
			mockSetup:    setupMocks(uploadParquetFile(errors.New("err")))(writeStructToParquet([]byte{}, nil)),
			taglist:      &entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: true,
			ctx:          ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup(t)
			repo, _ := NewTaglistStorage(logger, s3, parquet)
			err := repo.CreateTaglist(tt.ctx(t), tt.taglist)

			if tt.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadTaglist(t *testing.T) {
	tags := entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}}

	tests := []struct {
		name         string
		mockSetup    func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		taglist      *entity.TagListEntity
		ctx          func(t *testing.T) context.Context
		expectsError bool
	}{
		{
			name:         "happy path",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil))(readStructsFromParquet([]entity.TagListEntity{tags}, nil)),
			taglist:      &tags,
			expectsError: false,
			ctx:          ctx,
		},
		{
			name:         "nil ctx",
			mockSetup:    setupMocks()(),
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "s3 error",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, errors.New("err")))(),
			taglist:      nil,
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "parquet error",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil))(readStructsFromParquet([]entity.TagListEntity{}, errors.New("err"))),
			taglist:      nil,
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "invalid taglist returned",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil))(readStructsFromParquet([]entity.TagListEntity{{Tags: []string{"", ""}}}, nil)),
			taglist:      nil,
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "no taglist found",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil))(readStructsFromParquet([]entity.TagListEntity{}, nil)),
			taglist:      nil,
			expectsError: true,
			ctx:          ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup(t)
			repo, _ := NewTaglistStorage(logger, s3, parquet)
			taglist, err := repo.ReadTaglist(tt.ctx(t))

			if tt.expectsError {
				assert.Error(t, err)
				assert.Nil(t, taglist)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.taglist, taglist)
			}
		})
	}
}

func TestUpdateTaglist(t *testing.T) {
	tags := entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}}

	tests := []struct {
		name         string
		mockSetup    func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		taglist      *entity.TagListEntity
		ctx          func(t *testing.T) context.Context
		expectsError bool
	}{
		{
			name:         "happy path",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil), uploadParquetFile(nil))(writeStructToParquet([]byte{}, nil)),
			taglist:      &tags,
			expectsError: false,
			ctx:          ctx,
		},
		{
			name:         "ctx nil",
			mockSetup:    setupMocks()(),
			taglist:      &entity.TagListEntity{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "invalid taglist",
			mockSetup:    setupMocks()(),
			taglist:      &entity.TagListEntity{},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "s3 download error",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, errors.New("err")))(),
			taglist:      &tags,
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "parquet error",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil))(writeStructToParquet([]byte{}, errors.New("err"))),
			taglist:      &tags,
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "s3 upload error",
			mockSetup:    setupMocks(downloadParquetFile([]byte{}, map[string]string{}, nil), uploadParquetFile(errors.New("err")))(writeStructToParquet([]byte{}, nil)),
			taglist:      &tags,
			expectsError: true,
			ctx:          ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup(t)
			repo, _ := NewTaglistStorage(logger, s3, parquet)
			err := repo.UpdateTaglist(tt.ctx(t), tt.taglist)

			if tt.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTaglistExists(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		ctx            func(t *testing.T) context.Context
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "file exists",
			mockSetup:      setupMocks(fileExists(true, nil))(),
			ctx:            ctx,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "file doesn't exist",
			mockSetup:      setupMocks(fileExists(false, nil))(),
			ctx:            ctx,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:          "ctx nil",
			mockSetup:     setupMocks()(),
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup(t)
			repo, _ := NewTaglistStorage(logger, s3, parquet)
			exists, err := repo.TaglistExists(tt.ctx(t))

			if tt.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, exists)
			}
		})
	}
}

func TestNewTaglistStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name        string
		mockSetup   func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		logger      *slog.Logger
		expectError bool
	}{
		{
			name:        "happy path",
			mockSetup:   setupMocks()(),
			logger:      logger,
			expectError: false,
		},
		{
			name: "everything nil",
			mockSetup: func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
				return nil, nil
			},
			logger:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup(t)
			repo, err := NewTaglistStorage(tt.logger, s3, parquet)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}
