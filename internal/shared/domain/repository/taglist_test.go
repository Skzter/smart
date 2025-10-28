package repository

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
)

func TestCreateTaglist(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func() (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity])
		taglist      entity.TagListEntity
		expectsError bool
	}{
		{
			name: "no errors",
			mockSetup: func() (*mocks.MockS3StorageWrapper, *mocks.MockParquetFileWrapper[entity.TagListEntity]) {
				s3 := mocks.NewMockS3StorageWrapper(t)
				s3.On("UploadParquetFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				parquet := mocks.NewMockParquetFileWrapper[entity.TagListEntity](t)
				parquet.On("WriteStructToParquet", mock.Anything).Return([]byte{}, nil)
				return s3, parquet
			},
			taglist:      entity.TagListEntity{Tags: []string{"TAG1", "TAG2"}},
			expectsError: false,
		},
	}

	logger := slog.New(slog.DiscardHandler)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s3, parquet := tt.mockSetup()
			repo, _ := NewTaglistStorage(logger, s3, parquet)
			err := repo.CreateTaglist(t.Context(), tt.taglist)

			if tt.expectsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
