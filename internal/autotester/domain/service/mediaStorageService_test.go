package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

//nolint:dupl
func TestNewMediaStorageService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	mockRepo := mocks.NewMockMediaFileSystem(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.MediaFileSystem
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			repo:    mockRepo,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			repo:    mockRepo,
			wantErr: true,
		},
		{
			name:    "nil repo",
			logger:  logger,
			repo:    nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewMediaStorageService(test.logger, test.repo, tracer)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, svc, "service should be nil on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, svc, "service should not be nil")
			}
		})
	}
}

//nolint:dupl
func TestGetScreenshotUrl(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		testId      string
		setupMock   func(*mocks.MockMediaFileSystem)
		expectedUrl string
		wantErr     bool
	}{
		{
			name:   "success - screenshot url retrieved",
			testId: "test-123",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetScreenshotUrl(mock.Anything, "test-123").Return("https://s3.example.com/test-123.png", nil)
			},
			expectedUrl: "https://s3.example.com/test-123.png",
			wantErr:     false,
		},
		{
			name:   "error - screenshot not found",
			testId: "test-404",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetScreenshotUrl(mock.Anything, "test-404").Return("", errors.New("screenshot not found"))
			},
			expectedUrl: "",
			wantErr:     true,
		},
		{
			name:   "error - repository error",
			testId: "test-error",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetScreenshotUrl(mock.Anything, "test-error").Return("", errors.New("repo error"))
			},
			expectedUrl: "",
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockMediaFileSystem(t)

			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewMediaStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			url, err := svc.GetScreenshotUrl(ctx, test.testId)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Empty(t, url, "url should be empty on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, test.expectedUrl, url, "url should match expected")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

//nolint:dupl
func TestGetVideoUrl(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		testId      string
		setupMock   func(*mocks.MockMediaFileSystem)
		expectedUrl string
		wantErr     bool
	}{
		{
			name:   "success - video url retrieved",
			testId: "test-123",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetVideoUrl(mock.Anything, "test-123").Return("https://s3.example.com/test-123.webm", nil)
			},
			expectedUrl: "https://s3.example.com/test-123.webm",
			wantErr:     false,
		},
		{
			name:   "error - video not found",
			testId: "test-404",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetVideoUrl(mock.Anything, "test-404").Return("", errors.New("video not found"))
			},
			expectedUrl: "",
			wantErr:     true,
		},
		{
			name:   "error - repository error",
			testId: "test-error",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().GetVideoUrl(mock.Anything, "test-error").Return("", errors.New("repo error"))
			},
			expectedUrl: "",
			wantErr:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockMediaFileSystem(t)

			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewMediaStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			url, err := svc.GetVideoUrl(ctx, test.testId)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Empty(t, url, "url should be empty on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.Equal(t, test.expectedUrl, url, "url should match expected")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestHasMedia(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	tests := []struct {
		name               string
		testId             string
		setupMock          func(*mocks.MockMediaFileSystem)
		expectedScreenshot bool
		expectedVideo      bool
	}{
		{
			name:   "both screenshot and video exist",
			testId: "test-123",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().HasScreenshot(mock.Anything, "test-123").Return(true)
				mockRepo.EXPECT().HasVideo(mock.Anything, "test-123").Return(true)
			},
			expectedScreenshot: true,
			expectedVideo:      true,
		},
		{
			name:   "only screenshot exists",
			testId: "test-screenshot",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().HasScreenshot(mock.Anything, "test-screenshot").Return(true)
				mockRepo.EXPECT().HasVideo(mock.Anything, "test-screenshot").Return(false)
			},
			expectedScreenshot: true,
			expectedVideo:      false,
		},
		{
			name:   "only video exists",
			testId: "test-video",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().HasScreenshot(mock.Anything, "test-video").Return(false)
				mockRepo.EXPECT().HasVideo(mock.Anything, "test-video").Return(true)
			},
			expectedScreenshot: false,
			expectedVideo:      true,
		},
		{
			name:   "neither exists",
			testId: "test-none",
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().HasScreenshot(mock.Anything, "test-none").Return(false)
				mockRepo.EXPECT().HasVideo(mock.Anything, "test-none").Return(false)
			},
			expectedScreenshot: false,
			expectedVideo:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockMediaFileSystem(t)

			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewMediaStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			hasScreenshot, hasVideo := svc.HasMedia(ctx, test.testId)

			assert.Equal(t, test.expectedScreenshot, hasScreenshot, "hasScreenshot should match expected")
			assert.Equal(t, test.expectedVideo, hasVideo, "hasVideo should match expected")

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUploadMedia(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	testFile := entity.NewFile("screenshot.png", []byte("test data"), "png")

	tests := []struct {
		name      string
		testId    string
		file      entity.File
		setupMock func(*mocks.MockMediaFileSystem)
		wantErr   bool
	}{
		{
			name:   "success - media uploaded",
			testId: "test-123",
			file:   testFile,
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().UploadMedia(mock.Anything, "test-123", mock.MatchedBy(func(f entity.File) bool {
					return f.GetFileName() == "screenshot.png"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "error - upload fails",
			testId: "test-error",
			file:   testFile,
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().UploadMedia(mock.Anything, "test-error", mock.Anything).Return(errors.New("upload failed"))
			},
			wantErr: true,
		},
		{
			name:   "error - s3 error",
			testId: "test-s3-error",
			file:   testFile,
			setupMock: func(mockRepo *mocks.MockMediaFileSystem) {
				mockRepo.EXPECT().UploadMedia(mock.Anything, "test-s3-error", mock.Anything).Return(errors.New("s3 connection failed"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockMediaFileSystem(t)

			if test.setupMock != nil {
				test.setupMock(mockRepo)
			}

			svc, err := NewMediaStorageService(logger, mockRepo, tracer)
			require.NoError(t, err, "unexpected error creating service")

			err = svc.UploadMedia(ctx, test.testId, test.file)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
