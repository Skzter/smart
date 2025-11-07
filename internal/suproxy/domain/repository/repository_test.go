package repository

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// newTestLogger creates a new logger for testing purposes
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

const (
	testKey     = "tag1-tag2-123456"
	EntryPrefix = "supplierData/"
)

// getValidEntry returns a valid DatabaseEntry for testing
func getValidEntry() entity.DatabaseEntry {
	return entity.DatabaseEntry{
		Request:  "Test request",
		Response: entity.Response{Response: "OK"},
		Tags:     []string{"tag1", "tag2"},
	}
}

// TestNewDatabaseRepository tests the creation of a new database repository
func TestNewDatabaseRepository(t *testing.T) {
	_, mockS3, mockParquet := setupMocks(t)
	logger := newTestLogger()

	tests := []struct {
		name           string
		logger         *slog.Logger
		s3Wrapper      service.S3StorageWrapper
		parquetWrapper service.ParquetFileWrapper[entity.DatabaseEntry]
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
			repo, err := NewDatabaseRepository(test.logger, test.s3Wrapper, test.parquetWrapper, EntryPrefix)
			if (err != nil) != test.wantErr {
				t.Errorf("NewDatabaseRepository() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewDatabaseRepository() returned nil repository")
			}
		})
	}
}

// TestCreateRequest tests the CreateRequest method of the database repository
func TestCreateRequest(t *testing.T) {
	tests := []struct {
		name             string
		entry            entity.DatabaseEntry
		mockParquetError error
		mockS3Error      error
		expectedError    bool
	}{
		{
			name:          "success",
			entry:         getValidEntry(),
			expectedError: false,
		},
		{
			name:             "parquet error",
			entry:            getValidEntry(),
			mockParquetError: errors.New("parquet fail"),
			expectedError:    true,
		},
		{
			name:          "s3 upload error",
			entry:         getValidEntry(),
			mockS3Error:   errors.New("s3 fail"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)

			fakeData := []byte("parquet")
			mockParquet.On("WriteStructToParquet", tt.entry).Return(fakeData, tt.mockParquetError)

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

// TestReadRequest tests the ReadRequest method of the database repository
func TestReadRequest(t *testing.T) {
	tests := []struct {
		name                 string
		mockDownloadError    error
		mockParquetReadError error
		expectedError        bool
	}{
		{
			name:          "success",
			expectedError: false,
		},
		{
			name:              "download fails",
			mockDownloadError: errors.New("s3 error"),
			expectedError:     true,
		},
		{
			name:                 "parquet read fails",
			mockParquetReadError: errors.New("parquet read error"),
			expectedError:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)

			fakeData := []byte("parquet")
			metadata := map[string]string{"created": "123456"}

			mockS3.On("DownloadParquetFile", mock.Anything, EntryPrefix+testKey).Return(fakeData, metadata, tt.mockDownloadError)

			if tt.mockDownloadError == nil {
				mockParquet.On("ReadStructsFromParquet", fakeData).Return([]entity.DatabaseEntry{getValidEntry()}, tt.mockParquetReadError)
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

// TestUpdateRequest tests the UpdateRequest method of the database repository
func TestUpdateRequest(t *testing.T) {
	tests := []struct {
		name            string
		entry           entity.DatabaseEntry
		mockDownloadErr error
		mockParquetErr  error
		mockS3UploadErr error
		expectedError   bool
	}{
		{
			name:          "success",
			entry:         getValidEntry(),
			expectedError: false,
		},
		{
			name:            "download fails",
			entry:           getValidEntry(),
			mockDownloadErr: errors.New("download fail"),
			expectedError:   true,
		},
		{
			name:           "parquet write fails",
			entry:          getValidEntry(),
			mockParquetErr: errors.New("parquet fail"),
			expectedError:  true,
		},
		{
			name:            "s3 upload fails",
			entry:           getValidEntry(),
			mockS3UploadErr: errors.New("s3 upload fail"),
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, mockParquet := setupMocks(t)
			key := testKey
			fakeData := []byte("parquet")
			metadata := map[string]string{"created": "123456"}

			mockS3.On("DownloadParquetFile", mock.Anything, EntryPrefix+key).Return(fakeData, metadata, tt.mockDownloadErr)

			if tt.mockDownloadErr == nil {
				mockParquet.On("WriteStructToParquet", tt.entry).Return(fakeData, tt.mockParquetErr)
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

// TestDeleteRequest tests the DeleteRequest method of the database repository
func TestDeleteRequest(t *testing.T) {
	tests := []struct {
		name          string
		mockDeleteErr error
		expectedError bool
	}{
		{
			name:          "success",
			expectedError: false,
		},
		{
			name:          "delete fails",
			mockDeleteErr: errors.New("delete fail"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, _ := setupMocks(t)
			key := testKey
			mockS3.On("DeleteParquetFile", mock.Anything, EntryPrefix+key).Return(tt.mockDeleteErr)

			err := repo.DeleteRequest(context.Background(), key)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateDbEntry tests the validateDbEntry function
func TestValidateDbEntry(t *testing.T) {
	tests := []struct {
		name        string
		entry       entity.DatabaseEntry
		expectError bool
		errorText   string
	}{
		{
			name: "valid entry",
			entry: getValidEntry(),
			expectError: false,
		},

		{
			name: "invalid request - empty string",
			entry: entity.DatabaseEntry{
				Request:  "",
				Response: entity.Response{Response: "OK"},
				Tags:     []string{"tag1"},
			},
			expectError: true,
			errorText:   "request must not be empty",
		},
		{
			name: "invalid response - empty response",
			entry: entity.DatabaseEntry{
				Request:  "Test request",
				Response: entity.Response{Response: ""},
				Tags:     []string{"tag1"},
			},
			expectError: true,
			errorText:   "response must not be empty",
		},
		{
			name: "invalid tags - empty list",
			entry: entity.DatabaseEntry{
				Request:  "Test request",
				Response: entity.Response{Response: "OK"},
				Tags:     []string{},
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

// TestListAllKeys tests the ListAllKeys method of the database repository
func TestListAllKeys(t *testing.T) {
	tests := []struct {
		name         string
		mockKeys     []string
		mockErr      error
		expectedKeys []string
		expectErr    bool
	}{
		{
			name:         "success",
			mockKeys:     []string{"supplierData/tag1", "supplierData/tag2"},
			mockErr:      nil,
			expectedKeys: []string{"supplierData/tag1", "supplierData/tag2"},
			expectErr:    false,
		},
		{
			name:         "empty result",
			mockKeys:     []string{},
			mockErr:      nil,
			expectedKeys: []string{},
			expectErr:    false,
		},
		{
			name:         "s3 failure",
			mockKeys:     nil,
			mockErr:      errors.New("s3 error"),
			expectedKeys: nil,
			expectErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockS3, _ := setupMocks(t)

			mockS3.On("ListParquetFiles", mock.Anything, EntryPrefix).
				Return(tt.mockKeys, tt.mockErr).
				Once()

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

// setupMocks initializes the mocks for the database repository tests
func setupMocks(t *testing.T) (
	*databaseRepository,
	*mocks.MockS3StorageWrapper,
	*mocks.MockParquetFileWrapper[entity.DatabaseEntry],
) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)
	logger := newTestLogger()

	repo := &databaseRepository{
		s3Wrapper:      mockS3,
		parquetWrapper: mockParquet,
		logger:         logger,
		entryPrefix:    EntryPrefix,
	}

	return repo, mockS3, mockParquet
}
