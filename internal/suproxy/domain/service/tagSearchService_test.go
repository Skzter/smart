package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockWrapper "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
)

// TestNewTagSearchService tests the creation of a new TagSearchService instance.
func TestNewTagSearchService(t *testing.T) {
	mockS3 := mockWrapper.NewMockS3StorageWrapper(t)

	svc := NewTagSearchService(mockS3)

	assert.NotNil(t, svc)
}

// TestTagSearchService_FindKeysByTag tests the FindKeysByTag method of TagSearchService with different test cases.
func TestFindKeysByTag(t *testing.T) {
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
			mockKeys:     []string{"supplierData/2025-report.parquet", "backup/2025-summary.txt", "archive/2024.csv"},
			expectedKeys: nil,
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

			svc := &tagSearchService{
				s3: mockS3,
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

/*
	func extractKeysFromFile(parquetFile string) []string {
		// filename is: supplierData/no-hotelid_missing-tourdates.parquet
		// cuts suffix/prefix so only true filename remains
		ParquetFileNoSuffix, ok := strings.CutSuffix(parquetFile, ".parquet")
		if !ok {
			return nil
		}
		ParquetFileKeysOnly, ok := strings.CutPrefix(ParquetFileNoSuffix, "supplierData/")
		if !ok {
			return nil
		}
		// keys are seperated with "-" in filename
		keys := strings.Split(ParquetFileKeysOnly, "-")
		validKeys := []string{}
		for _, key := range keys {
			// sometimes number in filename, so if it errors its a string and true key
			if _, err := strconv.Atoi(key); err != nil {
				validKeys = append(validKeys, key)
			}
		}
		return validKeys
	}
*/
func TestExtractKeysFromFile(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		expectedKeys []string
	}{
		{
			name:         "wrong suffix - .BAM",
			filename:     "wrong.BAM",
			expectedKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keys := extractKeysFromFile(tc.filename)
			if !slices.Equal(keys, tc.expectedKeys) {
				t.Fatalf("slices dont equal:\nexpected => %v\ngot => %v\n", tc.expectedKeys, keys)
			}
		})
	}
}
