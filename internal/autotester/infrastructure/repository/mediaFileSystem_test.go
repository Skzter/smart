package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/wrapper"
)

func TestNewMediaFileSystem(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range newMediaFileSystemTestCaseProvider(t) {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewMediaFileSystem(test.s3Wrapper, tracer, test.prefix)
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func newMediaFileSystemTestCaseProvider(t *testing.T) []struct {
	name        string
	s3Wrapper   *mocks.MockS3StorageWrapper
	prefix      string
	expectError bool
} {
	return []struct {
		name        string
		s3Wrapper   *mocks.MockS3StorageWrapper
		prefix      string
		expectError bool
	}{
		{
			name:        "happy path with prefix",
			s3Wrapper:   mocks.NewMockS3StorageWrapper(t),
			prefix:      "media/test",
			expectError: false,
		},
		{
			name:        "no prefix",
			s3Wrapper:   mocks.NewMockS3StorageWrapper(t),
			prefix:      "",
			expectError: true,
		},
		{
			name:        "nil s3Wrapper",
			s3Wrapper:   nil,
			prefix:      "media",
			expectError: true,
		},
	}
}

func TestGetScreenshotUrl(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range getScreenshotUrlTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if test.getMediaUrlRet != nil {
				mockS3.On("GetMediaUrl", mock.Anything, test.expectedKey).Return(test.getMediaUrlRet...)
			}

			repo, _ := NewMediaFileSystem(mockS3, tracer, "media/screenshots")

			url, err := repo.GetScreenshotUrl(test.ctx, test.testId)
			if test.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedUrl, url)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func getScreenshotUrlTestCaseProvider() []struct {
	name           string
	testId         string
	expectedKey    string
	getMediaUrlRet []any
	expectedUrl    string
	expectError    bool
	ctx            context.Context
} {
	return getMediaUrlTestCases("media/screenshots", "test-123.png", "test-789.png")
}

func getMediaUrlTestCases(prefixPath, file1, file3 string) []struct {
	name           string
	testId         string
	expectedKey    string
	getMediaUrlRet []any
	expectedUrl    string
	expectError    bool
	ctx            context.Context
} {
	return []struct {
		name           string
		testId         string
		expectedKey    string
		getMediaUrlRet []any
		expectedUrl    string
		expectError    bool
		ctx            context.Context
	}{
		{
			name:           "happy path with prefix",
			testId:         "test-123",
			expectedKey:    prefixPath + "/" + file1,
			getMediaUrlRet: []any{"https://s3.example.com/" + file1, nil},
			expectedUrl:    "https://s3.example.com/" + file1,
			expectError:    false,
			ctx:            context.Background(),
		},
		{
			name:        "empty testId",
			testId:      "",
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:           "S3 error",
			testId:         "test-789",
			expectedKey:    prefixPath + "/" + file3,
			getMediaUrlRet: []any{"", errors.New("S3 error")},
			expectError:    true,
			ctx:            context.Background(),
		},
	}
}

func TestGetVideoUrl(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range getVideoUrlTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if test.getMediaUrlRet != nil {
				mockS3.On("GetMediaUrl", mock.Anything, test.expectedKey).Return(test.getMediaUrlRet...)
			}

			repo, _ := NewMediaFileSystem(mockS3, tracer, "media/videos")

			url, err := repo.GetVideoUrl(test.ctx, test.testId)
			if test.expectError {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedUrl, url)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func getVideoUrlTestCaseProvider() []struct {
	name           string
	testId         string
	expectedKey    string
	getMediaUrlRet []any
	expectedUrl    string
	expectError    bool
	ctx            context.Context
} {
	return getMediaUrlTestCases("media/videos", "test-123.webm", "test-789.webm")
}

func TestHasScreenshot(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range hasScreenshotTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if test.fileExistsRet != nil {
				mockS3.On("FileExists", mock.Anything, test.expectedKey).Return(test.fileExistsRet...)
			}

			repo, _ := NewMediaFileSystem(mockS3, tracer, "media/screenshots")

			exists := repo.HasScreenshot(test.ctx, test.testId)
			assert.Equal(t, test.expectedExists, exists)

			mockS3.AssertExpectations(t)
		})
	}
}

func hasScreenshotTestCaseProvider() []struct {
	name           string
	testId         string
	expectedKey    string
	fileExistsRet  []any
	expectedExists bool
	ctx            context.Context
} {
	return hasMediaTestCases("media/screenshots", "test-123.png", "test-789.png", "test-error.png", "screenshot")
}

func hasMediaTestCases(prefixPath, file1, file3, file4, mediaType string) []struct {
	name           string
	testId         string
	expectedKey    string
	fileExistsRet  []any
	expectedExists bool
	ctx            context.Context
} {
	return []struct {
		name           string
		testId         string
		expectedKey    string
		fileExistsRet  []any
		expectedExists bool
		ctx            context.Context
	}{
		{
			name:           mediaType + " exists with prefix",
			testId:         "test-123",
			expectedKey:    prefixPath + "/" + file1,
			fileExistsRet:  []any{true, nil},
			expectedExists: true,
			ctx:            context.Background(),
		},
		{
			name:           mediaType + " does not exist",
			testId:         "test-789",
			expectedKey:    prefixPath + "/" + file3,
			fileExistsRet:  []any{false, nil},
			expectedExists: false,
			ctx:            context.Background(),
		},
		{
			name:           "S3 error",
			testId:         "test-error",
			expectedKey:    prefixPath + "/" + file4,
			fileExistsRet:  []any{false, errors.New("S3 error")},
			expectedExists: false,
			ctx:            context.Background(),
		},
		{
			name:           "empty testId",
			testId:         "",
			expectedExists: false,
			ctx:            context.Background(),
		},
	}
}

func TestHasVideo(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range hasVideoTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if test.fileExistsRet != nil {
				mockS3.On("FileExists", mock.Anything, test.expectedKey).Return(test.fileExistsRet...)
			}

			repo, _ := NewMediaFileSystem(mockS3, tracer, "media/videos")

			exists := repo.HasVideo(test.ctx, test.testId)
			assert.Equal(t, test.expectedExists, exists)

			mockS3.AssertExpectations(t)
		})
	}
}

func hasVideoTestCaseProvider() []struct {
	name           string
	testId         string
	expectedKey    string
	fileExistsRet  []any
	expectedExists bool
	ctx            context.Context
} {
	return hasMediaTestCases("media/videos", "test-123.webm", "test-789.webm", "test-error.webm", "video")
}

func TestUploadMedia(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range uploadMediaTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)

			if test.uploadMediaFileRet != nil {
				mockS3.On("UploadMediaFile", mock.Anything, test.expectedKey, mock.Anything, mock.Anything).Return(test.uploadMediaFileRet...)
			}

			repo, _ := NewMediaFileSystem(mockS3, tracer, test.prefix)

			err := repo.UploadMedia(test.ctx, test.testId, test.file)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockS3.AssertExpectations(t)
		})
	}
}

func uploadMediaTestCaseProvider() []struct {
	name               string
	testId             string
	prefix             string
	file               entity.File
	expectedKey        string
	uploadMediaFileRet []any
	expectError        bool
	ctx                context.Context
} {
	return []struct {
		name               string
		testId             string
		prefix             string
		file               entity.File
		expectedKey        string
		uploadMediaFileRet []any
		expectError        bool
		ctx                context.Context
	}{
		{
			name:               "upload screenshot with prefix",
			testId:             "test-123",
			prefix:             "media/screenshots",
			file:               entity.NewFile("screenshot.png", []byte("image data"), "png"),
			expectedKey:        "media/screenshots/test-123.png",
			uploadMediaFileRet: []any{nil},
			expectError:        false,
			ctx:                context.Background(),
		},
		{
			name:               "upload video with prefix",
			testId:             "test-456",
			prefix:             "media/videos",
			file:               entity.NewFile("video.webm", []byte("video data"), "webm"),
			expectedKey:        "media/videos/test-456.webm",
			uploadMediaFileRet: []any{nil},
			expectError:        false,
			ctx:                context.Background(),
		},

		{
			name:        "empty testId",
			testId:      "",
			prefix:      "media",
			file:        entity.NewFile("screenshot.png", []byte("image data"), "png"),
			expectError: true,
			ctx:         context.Background(),
		},
		{
			name:               "S3 upload error",
			testId:             "test-error",
			prefix:             "media",
			file:               entity.NewFile("screenshot.png", []byte("image data"), "png"),
			expectedKey:        "media/test-error.png",
			uploadMediaFileRet: []any{errors.New("S3 upload failed")},
			expectError:        true,
			ctx:                context.Background(),
		},
	}
}

func TestGetMediaKey(t *testing.T) {
	tracer := otel.Tracer("test")
	for _, test := range getMediaKeyTestCaseProvider() {
		t.Run(test.name, func(t *testing.T) {
			mockS3 := mocks.NewMockS3StorageWrapper(t)
			fs, _ := NewMediaFileSystem(mockS3, tracer, test.prefix)

			// Type assert to access private method for testing
			mediaFs := fs.(*mediaFileSystem)
			key := mediaFs.getMediaKey(test.testId, test.extension)

			assert.Equal(t, test.expectedKey, key)
		})
	}
}

func getMediaKeyTestCaseProvider() []struct {
	name        string
	testId      string
	extension   string
	prefix      string
	expectedKey string
} {
	return []struct {
		name        string
		testId      string
		extension   string
		prefix      string
		expectedKey string
	}{
		{
			name:        "with prefix",
			testId:      "test-123",
			extension:   "png",
			prefix:      "media/screenshots",
			expectedKey: "media/screenshots/test-123.png",
		},
		{
			name:        "with different prefix",
			testId:      "test-456",
			extension:   "webm",
			prefix:      "media/videos",
			expectedKey: "media/videos/test-456.webm",
		},
		{
			name:        "nested prefix",
			testId:      "test-789",
			extension:   "png",
			prefix:      "media/test/screenshots",
			expectedKey: "media/test/screenshots/test-789.png",
		},
	}
}
