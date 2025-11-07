package service

import (
	"context"
	"errors"
	"slices"
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
				if err == nil {
					t.Fatal("wantend error but got nil")
				}
				if service != nil {
					t.Fatalf("wanted nil service, but got %v", service)
				}
			} else {
				assert.NotNil(t, service, err)
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
			expectedKeys: []string{"supplierData/2025-report.parquet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if tt.tag != "   " && tt.mockError != nil {
				mockS3.On("ListParquetFiles", mock.Anything, "").Return(nil, tt.mockError)
			} else if tt.tag != "   " {
				mockS3.On("ListParquetFiles", mock.Anything, "").Return(tt.mockKeys, nil)
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
	tests := []struct {
		name         string
		filename     string
		prefix       string
		expectedKeys []string
	}{
		{
			name:         "wrong suffix - .BAM",
			filename:     "wrong.BAM",
			prefix:       "supplierData/",
			expectedKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keys := extractKeysFromFile(tc.filename, tc.prefix)
			if !slices.Equal(keys, tc.expectedKeys) {
				t.Fatalf("slices dont equal:\nexpected => %v\ngot => %v\n", tc.expectedKeys, keys)
			}
		})
	}
}
