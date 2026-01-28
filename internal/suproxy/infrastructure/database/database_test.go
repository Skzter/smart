package database

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"

	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
)

func TestNewDatabaseRepository(t *testing.T) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	repo := NewDatabaseRepository(
		logger,
		mockS3,
		mockParquet,
		tracer,
		"prefix/",
	)

	assert.NotNil(t, repo)
}

func TestCreateRequest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	validEntry := entity.DatabaseEntry{
		Request: entity.Request{
			Header:      map[string]string{"h": "v"},
			Tags:        "tag",
			Destination: "dest",
			Body:        "body",
		},
		Response: entity.Response{Response: "ok"},
		Tags: &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{{Name: "x"}},
		},
	}

	tests := []struct {
		name        string
		entry       entity.DatabaseEntry
		writeRet    []any
		uploadRet   []any
		expectError bool
	}{
		{
			name:        "happy path",
			entry:       validEntry,
			writeRet:    []any{[]byte("parquet"), nil},
			uploadRet:   []any{nil},
			expectError: false,
		},
		{
			name:        "validation error",
			entry:       entity.DatabaseEntry{},
			expectError: true,
		},
		{
			name:        "parquet write error",
			entry:       validEntry,
			writeRet:    []any{[]byte(nil), errors.New("write failed")},
			expectError: true,
		},
		{
			name:        "s3 upload error",
			entry:       validEntry,
			writeRet:    []any{[]byte("parquet"), nil},
			uploadRet:   []any{errors.New("upload failed")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)

			repo := NewDatabaseRepository(
				logger,
				mockS3,
				mockParquet,
				tracer,
				"prefix/",
			)

			if tt.writeRet != nil {
				mockParquet.
					On("WriteStructToParquet", mock.Anything, tt.entry).
					Return(tt.writeRet...).
					Once()
			}

			if tt.uploadRet != nil {
				mockS3.
					On("UploadParquetFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).
					Return(tt.uploadRet...).
					Once()
			}

			err := repo.CreateRequest(context.Background(), tt.entry)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func TestReadRequest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	expected := entity.DatabaseEntry{
		Request: entity.Request{Tags: "abc"},
	}

	tests := []struct {
		name        string
		key         string
		downloadRet []any
		readRet     []any
		expectError bool
	}{
		{
			name:        "happy path",
			key:         "req-1",
			downloadRet: []any{[]byte("raw"), map[string]string{}, nil},
			readRet:     []any{[]entity.DatabaseEntry{expected}, nil},
			expectError: false,
		},
		{
			name:        "download error",
			key:         "req-1",
			downloadRet: []any{[]byte(nil), map[string]string{}, errors.New("download failed")},
			expectError: true,
		},
		{
			name:        "parquet read error",
			key:         "req-1",
			downloadRet: []any{[]byte("raw"), map[string]string{}, nil},
			readRet:     []any{[]entity.DatabaseEntry{}, errors.New("read failed")},
			expectError: true,
		},
		{
			name:        "empty key",
			key:         "",
			downloadRet: []any{[]byte(nil), map[string]string{}, errors.New("invalid key")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)

			repo := NewDatabaseRepository(
				logger,
				mockS3,
				mockParquet,
				tracer,
				"prefix/",
			)

			if tt.key != "" && tt.downloadRet != nil {
				mockS3.
					On("DownloadParquetFile", mock.Anything, "prefix/"+tt.key).
					Return(tt.downloadRet...).
					Once()
			}

			if tt.key == "" {
				mockS3.
					On("DownloadParquetFile", mock.Anything, "prefix/").
					Return(tt.downloadRet...).
					Once()
			}

			if tt.readRet != nil {
				mockParquet.
					On("ReadStructsFromParquet", mock.Anything, mock.Anything).
					Return(tt.readRet...).
					Once()
			}

			_, err := repo.ReadRequest(context.Background(), tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func TestUpdateRequest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	entry := entity.DatabaseEntry{
		Request: entity.Request{Tags: "abc"},
	}

	tests := []struct {
		name        string
		writeRet    []any
		uploadRet   []any
		expectError bool
	}{
		{
			name:        "happy path",
			writeRet:    []any{[]byte("data"), nil},
			uploadRet:   []any{nil},
			expectError: false,
		},
		{
			name:        "write error",
			writeRet:    []any{[]byte(nil), errors.New("write failed")},
			expectError: true,
		},
		{
			name:        "upload error",
			writeRet:    []any{[]byte("data"), nil},
			uploadRet:   []any{errors.New("upload failed")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)

			repo := NewDatabaseRepository(
				logger,
				mockS3,
				mockParquet,
				tracer,
				"prefix/",
			)

			mockParquet.
				On("WriteStructToParquet", mock.Anything, entry).
				Return(tt.writeRet...).
				Once()

			if tt.uploadRet != nil {
				mockS3.
					On("UploadParquetFile", mock.Anything, "prefix/key", mock.Anything, mock.Anything).
					Return(tt.uploadRet...).
					Once()
			}

			err := repo.UpdateRequest(context.Background(), "key", entry)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
			mockParquet.AssertExpectations(t)
		})
	}
}

func TestDeleteRequest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		ctx         context.Context
		key         string
		deleteRet   []any
		expectError bool
	}{
		{
			name:        "happy path",
			ctx:         context.Background(),
			key:         "req-1",
			deleteRet:   []any{nil},
			expectError: false,
		},
		{
			name:        "nil ctx",
			ctx:         nil,
			key:         "req-1",
			deleteRet:   []any{errors.New("invalid ctx")},
			expectError: true,
		},
		{
			name:        "empty key",
			ctx:         context.Background(),
			key:         "",
			deleteRet:   []any{errors.New("invalid key")},
			expectError: true,
		},
		{
			name:        "s3 delete error",
			ctx:         context.Background(),
			key:         "req-1",
			deleteRet:   []any{errors.New("delete failed")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)

			repo := NewDatabaseRepository(
				logger,
				mockS3,
				mockParquet,
				tracer,
				"prefix/",
			)

			mockS3.On("DeleteParquetFile", tt.ctx, "prefix/"+tt.key).
				Return(tt.deleteRet...).
				Once()

			err := repo.DeleteRequest(tt.ctx, tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func TestListAllKeys(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		listRet     []any
		expectError bool
	}{
		{
			name:        "happy path",
			listRet:     []any{[]string{"a", "b"}, nil},
			expectError: false,
		},
		{
			name:        "s3 error",
			listRet:     []any{[]string(nil), errors.New("list failed")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)

			repo := NewDatabaseRepository(
				logger,
				mockS3,
				mockParquet,
				tracer,
				"prefix/",
			)

			mockS3.
				On("ListParquetFiles", mock.Anything, "prefix/").
				Return(tt.listRet...).
				Once()

			_, err := repo.ListAllKeys(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
		})
	}
}
