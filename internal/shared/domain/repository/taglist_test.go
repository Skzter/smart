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

func ctx(t *testing.T) context.Context { return t.Context() }

func TestCreateTaglist(t *testing.T) {
	tests := []struct {
		name          string
		uploadReturns *[]any
		writeReturns  *[]any
		taglist       *entity.TagList
		ctx           func(t *testing.T) context.Context
		expectsError  bool
	}{
		{
			name:          "happy path",
			uploadReturns: &[]any{nil},
			writeReturns:  &[]any{[]byte{}, nil},
			taglist:       &entity.TagList{Tags: []string{"TAG1", "TAG2"}},
			expectsError:  false,
			ctx:           ctx,
		},
		{
			name:         "nil ctx",
			taglist:      &entity.TagList{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "validation error",
			taglist:      &entity.TagList{},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:         "parquet error",
			writeReturns: &[]any{nil, errors.New("err")},
			taglist:      &entity.TagList{Tags: []string{"TAG1", "TAG2"}},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:          "s3 error",
			uploadReturns: &[]any{errors.New("err")},
			writeReturns:  &[]any{[]byte{}, nil},
			taglist:       &entity.TagList{Tags: []string{"TAG1", "TAG2"}},
			expectsError:  true,
			ctx:           ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3 := mocks.NewMockS3StorageWrapper(t)
			parquet := mocks.NewMockParquetFileWrapper[entity.TagList](t)

			if tt.uploadReturns != nil {
				s3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(*tt.uploadReturns...)
			}

			if tt.writeReturns != nil {
				parquet.On("WriteStructToParquet", mock.Anything).Return(*tt.writeReturns...)
			}

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
	tags := entity.TagList{Tags: []string{"TAG1", "TAG2"}}

	tests := []struct {
		name            string
		downloadReturns *[]any
		readReturns     *[]any
		taglist         *entity.TagList
		ctx             func(t *testing.T) context.Context
		expectsError    bool
	}{
		{
			name:            "happy path",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			readReturns:     &[]any{[]entity.TagList{tags}, nil},
			taglist:         &tags,
			expectsError:    false,
			ctx:             ctx,
		},
		{
			name:         "nil ctx",
			taglist:      nil,
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:            "s3 error",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, errors.New("err")},
			taglist:         nil,
			expectsError:    true,
			ctx:             ctx,
		},
		{
			name:            "parquet error",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			readReturns:     &[]any{[]entity.TagList{}, errors.New("err")},
			taglist:         nil,
			expectsError:    true,
			ctx:             ctx,
		},
		{
			name:            "invalid taglist returned",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			readReturns:     &[]any{[]entity.TagList{{Tags: []string{"", ""}}}, nil},
			taglist:         nil,
			expectsError:    true,
			ctx:             ctx,
		},
		{
			name:            "no taglist found",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			readReturns:     &[]any{[]entity.TagList{}, nil},
			taglist:         nil,
			expectsError:    true,
			ctx:             ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3 := mocks.NewMockS3StorageWrapper(t)
			parquet := mocks.NewMockParquetFileWrapper[entity.TagList](t)

			if tt.downloadReturns != nil {
				s3.On("DownloadParquetFile", mock.Anything, mock.Anything).Return(*tt.downloadReturns...)
			}

			if tt.readReturns != nil {
				parquet.On("ReadStructsFromParquet", mock.Anything).Return(*tt.readReturns...)
			}

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
	tags := entity.TagList{Tags: []string{"TAG1", "TAG2"}}

	tests := []struct {
		name            string
		uploadReturns   *[]any
		downloadReturns *[]any
		writeReturns    *[]any
		taglist         *entity.TagList
		ctx             func(t *testing.T) context.Context
		expectsError    bool
	}{
		{
			name:            "happy path",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			uploadReturns:   &[]any{nil},
			writeReturns:    &[]any{[]byte{}, nil},
			taglist:         &tags,
			expectsError:    false,
			ctx:             ctx,
		},
		{
			name:         "ctx nil",
			taglist:      &entity.TagList{},
			expectsError: true,
			ctx:          func(t *testing.T) context.Context { return nil },
		},
		{
			name:         "invalid taglist",
			taglist:      &entity.TagList{},
			expectsError: true,
			ctx:          ctx,
		},
		{
			name:            "s3 download error",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, errors.New("err")},
			taglist:         &tags,
			expectsError:    true,
			ctx:             ctx,
		},
		{
			name:            "parquet error",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			writeReturns:    &[]any{[]byte{}, errors.New("err")},
			taglist:         &tags,
			expectsError:    true,
			ctx:             ctx,
		},
		{
			name:            "s3 upload error",
			downloadReturns: &[]any{[]byte{}, map[string]string{}, nil},
			uploadReturns:   &[]any{errors.New("err")},
			writeReturns:    &[]any{[]byte{}, nil},
			taglist:         &tags,
			expectsError:    true,
			ctx:             ctx,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3 := mocks.NewMockS3StorageWrapper(t)
			parquet := mocks.NewMockParquetFileWrapper[entity.TagList](t)

			if tt.uploadReturns != nil {
				s3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(*tt.uploadReturns...)
			}

			if tt.writeReturns != nil {
				parquet.On("WriteStructToParquet", mock.Anything).Return(*tt.writeReturns...)
			}

			if tt.downloadReturns != nil {
				s3.On("DownloadParquetFile", mock.Anything, mock.Anything).Return(*tt.downloadReturns...)
			}

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
		existReturns   *[]any
		ctx            func(t *testing.T) context.Context
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "file exists",
			existReturns:   &[]any{true, nil},
			ctx:            ctx,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "file doesn't exist",
			existReturns:   &[]any{false, nil},
			ctx:            ctx,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:          "ctx nil",
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3 := mocks.NewMockS3StorageWrapper(t)
			parquet := mocks.NewMockParquetFileWrapper[entity.TagList](t)

			if tt.existReturns != nil {
				s3.On("FileExists", mock.Anything, mock.Anything).Return(*tt.existReturns...)
			}

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
		mockSetup   func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagList])
		logger      *slog.Logger
		expectError bool
	}{
		{
			name: "happy path",
			mockSetup: func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagList]) {
				return mocks.NewMockS3StorageWrapper(t), mocks.NewMockParquetFileWrapper[entity.TagList](t)
			},
			logger:      logger,
			expectError: false,
		},
		{
			name: "everything nil",
			mockSetup: func(t *testing.T) (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagList]) {
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
