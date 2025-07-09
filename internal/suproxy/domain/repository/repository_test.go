package repository

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

func getValidEntry() entity.DatabaseEntry {
	return entity.DatabaseEntry{
		Request: entity.Request{
			Header:      []string{"Content-Type: application/json"},
			Prompt:      "prompt",
			Destination: "http://example.com",
			Request:     `{}`,
		},
		Response: entity.Response{Response: "OK"},
		Tags:     []string{"tag1", "tag2"},
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
			repo, mockS3, mockParquet, _ := setupMocks(t)

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

// setupMocks initializes the mocks for the database repository tests
func setupMocks(t *testing.T) (*databaseRepository, *mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.DatabaseEntry], *slog.Logger) {
	mockS3 := mocks.NewMockS3StorageWrapper(t)
	mockParquet := mocks.NewMockParquetFileWrapper[entity.DatabaseEntry](t)
	logger := slog.Default()

	repo := &databaseRepository{
		s3Wrapper:      mockS3,
		parquetWrapper: mockParquet,
		logger:         logger,
	}
	return repo, mockS3, mockParquet, logger
}
