package repository

import (
	"context"
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	wrapperService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/wrapper"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

const (
	screenshotExtension = "png"
	videoExtension      = "webm"
)

// mediaFileSystem is an implementation of the MediaFileSystem interface using S3.
type mediaFileSystem struct {
	s3Wrapper wrapperService.S3StorageWrapper
	prefix    string
}

// NewMediaFileSystem returns a MediaFileSystem that uses S3 for storage.
// The prefix is prepended to all S3 keys.
func NewMediaFileSystem(s3Wrapper wrapperService.S3StorageWrapper, prefix string) (repository.MediaFileSystem, error) {
	if err := assert.NotNil(s3Wrapper); err != nil {
		return nil, fmt.Errorf("s3Wrapper must not be nil")
	}

	return &mediaFileSystem{
		s3Wrapper: s3Wrapper,
		prefix:    prefix,
	}, nil
}

func (fs *mediaFileSystem) getMediaKey(testId, extension string) string {
	if fs.prefix != "" {
		return fmt.Sprintf("%s/%s.%s", fs.prefix, testId, extension)
	}
	return fmt.Sprintf("%s.%s", testId, extension)
}

func (fs *mediaFileSystem) GetScreenshotUrl(ctx context.Context, testId string) (string, error) {
	if err := assert.StringNotEmpty(testId); err != nil {
		return "", fmt.Errorf("testId must not be empty")
	}

	key := fs.getMediaKey(testId, screenshotExtension)
	return fs.s3Wrapper.GetMediaUrl(ctx, key)
}

func (fs *mediaFileSystem) GetVideoUrl(ctx context.Context, testId string) (string, error) {
	if err := assert.StringNotEmpty(testId); err != nil {
		return "", fmt.Errorf("testId must not be empty")
	}

	key := fs.getMediaKey(testId, videoExtension)
	return fs.s3Wrapper.GetMediaUrl(ctx, key)
}

func (fs *mediaFileSystem) HasScreenshot(ctx context.Context, testId string) bool {
	if err := assert.StringNotEmpty(testId); err != nil {
		return false
	}

	key := fs.getMediaKey(testId, screenshotExtension)
	exists, err := fs.s3Wrapper.FileExists(ctx, key)
	return err == nil && exists
}

func (fs *mediaFileSystem) HasVideo(ctx context.Context, testId string) bool {
	if err := assert.StringNotEmpty(testId); err != nil {
		return false
	}

	key := fs.getMediaKey(testId, videoExtension)
	exists, err := fs.s3Wrapper.FileExists(ctx, key)
	return err == nil && exists
}

func (fs *mediaFileSystem) UploadMedia(ctx context.Context, testId string, file entity.File) error {
	if err := assert.StringNotEmpty(testId); err != nil {
		return fmt.Errorf("testId must not be empty")
	}

	// Upload to S3
	key := fs.getMediaKey(testId, file.GetFileExtension())
	if err := fs.s3Wrapper.UploadMediaFile(ctx, key, file.GetFileData(), nil); err != nil {
		return fmt.Errorf("failed to upload %s to S3: %w", file.GetFileExtension(), err)
	}

	return nil
}
