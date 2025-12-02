package repository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	wrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const (
	testKey     = "tag1-tag2-123456"
	EntryPrefix = "supplierData/"
)

func getValidEntry() entity.DatabaseEntry {
	return entity.DatabaseEntry{
		Request: entity.Request{
			Header:      map[string]string{"Content-Type": "application/json"},
			Tags:        "Tags",
			Destination: "http://example.com",
			Body:        `{}`,
		},
		Response: entity.Response{Response: "OK"},
		Tags: &sharedEntity.TagList{
			Tags: []sharedEntity.Tag{
				{Name: "tag1"},
				{Name: "tag2"},
			},
		},
	}
}

func TestNewDatabaseRepository(t *testing.T) {
	tracer := otel.Tracer("test")
	_, mockS3, mockParquet := setupMocks(t)
	logger := newTestLogger()

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry]
		wantErr        bool
	}{
		{"all not nil", logger, mockS3, mockParquet, false},
		{"nil logger", nil, mockS3, mockParquet, true},
		{"nil s3Wrapper", logger, nil, mockParquet, true},
		{"nil parquetWrapper", logger, mockS3, nil, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewDatabaseRepository(test.logger, test.s3Wrapper, test.parquetWrapper, tracer, EntryPrefix)
			if (err != nil) != test.wantErr {
				t.Errorf("NewDatabaseRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewDatabaseRepository() returned nil repository")
			}
		})
	}
}

func TestCreateRequest(t *testing.T) {
	tests := []struct {
		name             string
		entry            entity.DatabaseEntry
		mockParquetError error
		mockS3Error      error
		expectedError    bool
	}{
		{"success", getValidEntry(), nil, nil, false},
		{"parquet error", getValidEntry(), errors.New("parquet fail"), nil, true},
		{"s3 upload error", getValidEntry(), nil, errors.New("s3 fail"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)
			fakeData := []byte("parquet")

			mockParquet.On("WriteStructToParquet", mock.Anything, tt.entry).Return(fakeData, tt.mockParquetError)
			if tt.mockParquetError == nil {
				mockS3.On("UploadParquetFile", mock.Anything, mock.AnythingOfType("string"), fakeData, mock.Anything).Return(tt.mockS3Error)
			}

			err := repo.CreateRequest(context.Background(), tt.entry)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadRequest(t *testing.T) {
	tests := []struct {
		name                 string
		mockDownloadError    error
		mockParquetReadError error
		expectedError        bool
	}{
		{"success", nil, nil, false},
		{"download fails", errors.New("s3 error"), nil, true},
		{"parquet read fails", nil, errors.New("parquet read error"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)
			fakeData := []byte("parquet")
			metadata := map[string]string{"created": "123456"}

			mockS3.On("DownloadParquetFile", mock.Anything, EntryPrefix+testKey).Return(fakeData, metadata, tt.mockDownloadError)
			if tt.mockDownloadError == nil {
				mockParquet.On("ReadStructsFromParquet", mock.Anything, fakeData).Return([]entity.DatabaseEntry{getValidEntry()}, tt.mockParquetReadError)
			}

			_, err := repo.ReadRequest(context.Background(), testKey)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateRequest(t *testing.T) {
	tests := []struct {
		name            string
		entry           entity.DatabaseEntry
		mockDownloadErr error
		mockParquetErr  error
		mockS3UploadErr error
		expectedError   bool
	}{
		{"success", getValidEntry(), nil, nil, nil, false},
		{"download fails", getValidEntry(), errors.New("download fail"), nil, nil, true},
		{"parquet write fails", getValidEntry(), nil, errors.New("parquet fail"), nil, true},
		{"s3 upload fails", getValidEntry(), nil, nil, errors.New("s3 upload fail"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)
			key := testKey
			fakeData := []byte("parquet")
			metadata := map[string]string{"created": "123456"}

			mockS3.On("DownloadParquetFile", mock.Anything, EntryPrefix+key).Return(fakeData, metadata, tt.mockDownloadErr)
			if tt.mockDownloadErr == nil {
				mockParquet.On("WriteStructToParquet", mock.Anything, tt.entry).Return(fakeData, tt.mockParquetErr)
			}
			if tt.mockDownloadErr == nil && tt.mockParquetErr == nil {
				mockS3.On("UploadParquetFile", mock.Anything, EntryPrefix+key, fakeData, mock.Anything).Return(tt.mockS3UploadErr)
			}

			err := repo.UpdateRequest(context.Background(), key, tt.entry)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteRequest(t *testing.T) {
	tests := []struct {
		name          string
		mockDeleteErr error
		expectedError bool
	}{
		{"success", nil, false},
		{"delete fails", errors.New("delete fail"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, _ := setupMocks(t)
			mockS3.On("DeleteParquetFile", mock.Anything, EntryPrefix+testKey).Return(tt.mockDeleteErr)

			err := repo.DeleteRequest(context.Background(), testKey)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDbEntry(t *testing.T) {
	tests := []struct {
		name        string
		entry       entity.DatabaseEntry
		expectError bool
		errorText   string
	}{
		{
			name: "valid entry",
			entry: entity.DatabaseEntry{
				Request: entity.Request{
					Header:      map[string]string{"Content-Type": "application/json"},
					Tags:        "Tags",
					Destination: "http://example.com",
					Body:        "{}",
				},
				Response: entity.Response{Response: "OK"},
				Tags:     getValidEntry().Tags,
			},
			expectError: false,
		},
		{
			name: "invalid request - empty header",
			entry: entity.DatabaseEntry{
				Request: entity.Request{
					Header:      map[string]string{},
					Tags:        "Tags",
					Destination: "http://example.com",
					Body:        "{}",
				},
				Response: entity.Response{Response: "OK"},
				Tags:     getValidEntry().Tags,
			},
			expectError: true,
			errorText:   "header must not be empty",
		},

		{
			name: "invalid request - empty tags",
			entry: entity.DatabaseEntry{
				Request: entity.Request{
					Header:      map[string]string{"Content-Type": "application/json"},
					Tags:        "",
					Destination: "http://example.com",
					Body:        "{}",
				},
				Response: entity.Response{Response: "OK"},
				Tags:     getValidEntry().Tags,
			},
			expectError: true,
			errorText:   "tags must not be empty",
		},
		{
			name: "invalid tags - nil",
			entry: entity.DatabaseEntry{
				Request:  getValidEntry().Request,
				Response: getValidEntry().Response,
				Tags:     nil,
			},
			expectError: true,
			errorText:   "tags must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDbEntry(tt.entry)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListAllKeys(t *testing.T) {
	tests := []struct {
		name         string
		mockKeys     []string
		mockErr      error
		expectedKeys []string
		expectErr    bool
	}{
		{"success", []string{"supplierData/tag1", "supplierData/tag2"}, nil, []string{"supplierData/tag1", "supplierData/tag2"}, false},
		{"empty result", []string{}, nil, []string{}, false},
		{"s3 failure", nil, errors.New("s3 error"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, _ := setupMocks(t)
			mockS3.On("ListParquetFiles", mock.Anything, EntryPrefix).Return(tt.mockKeys, tt.mockErr).Once()

			keys, err := repo.ListAllKeys(context.Background())
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedKeys, keys)
			}
			mockS3.AssertExpectations(t)
		})
	}
}

func setupMocks(t *testing.T) (
	*databaseRepository,
	*wrapper.MockS3StorageWrapper,
	*wrapper.MockParquetFileWrapper[entity.DatabaseEntry],
) {
	mockS3 := wrapper.NewMockS3StorageWrapper(t)
	mockParquet := wrapper.NewMockParquetFileWrapper[entity.DatabaseEntry](t)
	logger := newTestLogger()
	tracer := otel.Tracer("test")

	repo := &databaseRepository{
		s3Wrapper:      mockS3,
		parquetWrapper: mockParquet,
		logger:         logger,
		tracer:         tracer,
		entryPrefix:    EntryPrefix,
	}
	return repo, mockS3, mockParquet
}
