package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockWrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

// TestNewTagSearchService tests the creation of a new TagSearchService instance.
func TestNewTagSearchService(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockS3 := mockWrapper.NewMockS3StorageWrapper(t)

	// Dereference logger pointer to pass slog.Logger value
	svc := NewTagSearchService(logger, mockS3)

	assert.NotNil(t, svc)
}

// TestTagSearchService_FindKeysByTag tests the FindKeysByTag method of TagSearchService with different test cases.
func TestFindKeysByTag(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name         string
		tag          string
		mockKeys     []string
		mockError    error
		expectedKeys []string
		expectError  bool
	}{
		{
			name:        "empty tag",
			tag:         "   ",
			expectError: true,
		},
		{
			name:        "S3 error",
			tag:         "data",
			mockError:   errors.New("s3 unavailable"),
			expectError: true,
		},
		{
			name:         "no matches",
			tag:          "2025",
			mockKeys:     []string{"file1.parquet", "something.txt"},
			expectedKeys: []string(nil),
		},
		{
			name:         "some matches",
			tag:          "2025",
			mockKeys:     []string{"data/2025-report.parquet", "backup/2025-summary.txt", "archive/2024.csv"},
			expectedKeys: []string{"data/2025-report.parquet", "backup/2025-summary.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mockWrapper.NewMockS3StorageWrapper(t)

			if tt.tag != "   " && tt.mockError != nil {
				mockS3.On("ListParquetFiles", mock.Anything, "").Return(nil, tt.mockError)
			} else if tt.tag != "   " {
				mockS3.On("ListParquetFiles", mock.Anything, "").Return(tt.mockKeys, nil)
			}

			// Dereference logger pointer to pass slog.Logger value
			svc := &tagSearchService{
				logger: logger,
				s3:     mockS3,
			}

			keys, err := svc.FindKeysByTag(context.Background(), tt.tag)

			if tt.expectError {
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
