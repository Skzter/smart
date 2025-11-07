package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
)

// TestNewTagSearchService tests the creation of a new TagSearchService instance.
func TestNewTagSearchService(t *testing.T) {
	cfg, _ := config.LoadAppConfig()
	tests := []struct {
		name    string
		s3      service.S3StorageWrapper
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "valid service",
			s3:      mocks.NewMockS3StorageWrapper(t),
			cfg:     cfg,
			wantErr: false,
		},
		{
			name:    "invalid service",
			s3:      nil,
			cfg:     nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewTagSearchService(tc.cfg, tc.s3)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, service)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

// TestTagSearchService_FindKeysByTag tests the FindKeysByTag method of TagSearchService with different test cases.
func TestFindKeysByTag(t *testing.T) {
	cfg, _ := config.LoadAppConfig()
	tests := []struct {
		name         string
		tag          string
		mockResponse []any
		expectedKeys []string
		expectError  bool
	}{
		{
			name:        "empty tag",
			tag:         "   ",
			expectError: true,
		},
		{
			name:         "S3 error",
			tag:          "data",
			mockResponse: []any{nil, errors.New("s3 unavailable")},
			expectedKeys: nil,
			expectError:  true,
		},
		{
			name:         "no matches",
			tag:          "2025",
			mockResponse: []any{[]string{"data/file1.parquet", "something.txt"}, nil},
			expectedKeys: nil,
			expectError:  false,
		},
		{
			name:         "some matches",
			tag:          "no_hotelid",
			mockResponse: []any{[]string{"supplierData/no_hotelid-missing_tourdates.parquet", "backup/2025-summary.txt", "archive/2024.csv"}, nil},
			expectedKeys: []string{"supplierData/no_hotelid-missing_tourdates.parquet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if tt.mockResponse != nil {
				mockS3.On("ListParquetFiles", mock.Anything, mock.Anything).Return(tt.mockResponse...)
			}

			svc := &tagSearchService{
				config: cfg,
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

func TestExtractKeysFromFile(t *testing.T) {
	goodPrefix := "supplierData/"
	tests := []struct {
		name         string
		filename     string
		expectedKeys []string
	}{
		{
			name:         "valid file",
			filename:     "supplierData/no_hotelid-missing_destination.parquet",
			expectedKeys: []string{"no_hotelid", "missing_destination"},
		},
		{
			name:         "valid file with number at the end",
			filename:     "supplierData/no_hotelid-missing_destination-912830913.parquet",
			expectedKeys: []string{"no_hotelid", "missing_destination"},
		},
		{
			name:         "wrong suffix - .BAM",
			filename:     "supplierData/wrong.BAM",
			expectedKeys: nil,
		},
		{
			name:         "wrong prefix - shababs/",
			filename:     "shababs/wrong.parquet",
			expectedKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keys := extractKeysFromFile(tc.filename, goodPrefix)
			assert.Equal(t, tc.expectedKeys, keys)
		})
	}
}

func TestIsKeysInTag(t *testing.T) {
	tests := []struct {
		name     string
		keys     []string
		tag      string
		wantTrue bool
	}{
		{
			name:     "keys in tag",
			keys:     []string{"tag1", "tag2", "tag3"},
			tag:      "tag1 tag3",
			wantTrue: true,
		},
		{
			name:     "keys not in tag",
			keys:     []string{"tag1", "tag2", "tag3"},
			tag:      "hier ist nichts drinnen",
			wantTrue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok := isKeysInTag(tc.keys, tc.tag)
			if tc.wantTrue {
				assert.True(t, ok)
			} else {
				assert.False(t, ok)
			}
		})
	}
}
