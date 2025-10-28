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

func arg(name string, params int, returns ...any) mockCall {
	return mockCall{Name: name, Params: params, Returns: returns}
}

type mockCall struct {
	Name    string
	Params  int
	Returns []any
}

func setupMocks(s3calls ...mockCall) func(...mockCall) func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
	return func(parquetCalls ...mockCall) func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
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
		taglist      entity.TagListEntity
		ctx          func(t *testing.T) context.Context
		expectsError bool
	}{
		{
			name:         "happy path",
			mockSetup:    setupMocks(arg("UploadParquetFile", 4, nil))(arg("WriteStructToParquet", 1, []byte{}, nil)),
			taglist:      entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: false,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "nil ctx",
			mockSetup:    setupMocks()(),
			taglist:      entity.TagListEntity{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "validation error",
			mockSetup:    setupMocks()(),
			taglist:      entity.TagListEntity{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "parquet error",
			mockSetup:    setupMocks()(arg("WriteStructToParquet", 1, nil, errors.New("err"))),
			taglist:      entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "s3 error",
			mockSetup:    setupMocks(arg("UploadParquetFile", 4, errors.New("err")))(arg("WriteStructToParquet", 1, []byte{}, nil)),
			taglist:      entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
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
			mockSetup:    setupMocks(arg("DownloadParquetFile", 2, []byte{}, map[string]string{}, nil))(arg("ReadStructsFromParquet", 1, []entity.TagListEntity{tags}, nil)),
			taglist:      &tags,
			expectsError: false,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
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
			mockSetup:    setupMocks(arg("DownloadParquetFile", 2, []byte{}, map[string]string{}, errors.New("err")))(),
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "parquet error",
			mockSetup:    setupMocks(arg("DownloadParquetFile", 2, []byte{}, map[string]string{}, nil))(arg("ReadStructsFromParquet", 1, []entity.TagListEntity{}, errors.New("err"))),
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "invalid taglist returned",
			mockSetup:    setupMocks(arg("DownloadParquetFile", 2, []byte{}, map[string]string{}, nil))(arg("ReadStructsFromParquet", 1, []entity.TagListEntity{{Tags: []string{"", ""}}}, nil)),
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
		},
		{
			name:         "no taglist found",
			mockSetup:    setupMocks(arg("DownloadParquetFile", 2, []byte{}, map[string]string{}, nil))(arg("ReadStructsFromParquet", 1, []entity.TagListEntity{}, nil)),
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return t.Context() },
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
