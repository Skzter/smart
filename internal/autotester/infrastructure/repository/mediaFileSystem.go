package repository

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	tracer    trace.Tracer
}

// NewMediaFileSystem returns a MediaFileSystem that uses S3 for storage.
// The prefix is prepended to all S3 keys.
func NewMediaFileSystem(s3Wrapper wrapperService.S3StorageWrapper, tracer trace.Tracer, prefix string) (repository.MediaFileSystem, error) {
	if err := assert.NotNil(s3Wrapper, tracer); err != nil {
		return nil, err
	}

	if err := assert.StringNotEmpty(prefix); err != nil {
		return nil, err
	}

	return &mediaFileSystem{
		s3Wrapper: s3Wrapper,
		prefix:    prefix,
		tracer:    tracer,
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

	ctx, span := fs.tracer.Start(ctx, "mediaFileSystemRepository.GetScreenshotUrl")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", testId),
		attribute.String("media.type", "screenshot"),
	)

	key := fs.getMediaKey(testId, screenshotExtension)
	url, err := fs.s3Wrapper.GetMediaUrl(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get screenshot URL")
		return "", err
	}

	span.AddEvent("screenshot URL retrieved", trace.WithAttributes(
		attribute.String("media.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return url, nil
}

func (fs *mediaFileSystem) GetVideoUrl(ctx context.Context, testId string) (string, error) {
	if err := assert.StringNotEmpty(testId); err != nil {
		return "", fmt.Errorf("testId must not be empty")
	}

	ctx, span := fs.tracer.Start(ctx, "mediaFileSystemRepository.GetVideoUrl")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", testId),
		attribute.String("media.type", "video"),
	)

	key := fs.getMediaKey(testId, videoExtension)
	url, err := fs.s3Wrapper.GetMediaUrl(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get video URL")
		return "", err
	}

	span.AddEvent("video URL retrieved", trace.WithAttributes(
		attribute.String("media.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return url, nil
}

func (fs *mediaFileSystem) HasScreenshot(ctx context.Context, testId string) bool {
	if err := assert.StringNotEmpty(testId); err != nil {
		return false
	}

	ctx, span := fs.tracer.Start(ctx, "mediaFileSystemRepository.HasScreenshot")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", testId),
		attribute.String("media.type", "screenshot"),
	)

	key := fs.getMediaKey(testId, screenshotExtension)
	exists, err := fs.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check screenshot existence")
		return false
	}

	span.AddEvent("screenshot existence checked", trace.WithAttributes(
		attribute.String("media.key", key),
		attribute.Bool("media.exists", exists),
	))
	span.SetStatus(codes.Ok, "")

	return exists
}

func (fs *mediaFileSystem) HasVideo(ctx context.Context, testId string) bool {
	if err := assert.StringNotEmpty(testId); err != nil {
		return false
	}

	ctx, span := fs.tracer.Start(ctx, "mediaFileSystemRepository.HasVideo")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", testId),
		attribute.String("media.type", "video"),
	)

	key := fs.getMediaKey(testId, videoExtension)
	exists, err := fs.s3Wrapper.FileExists(ctx, key)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to check video existence")
		return false
	}

	span.AddEvent("video existence checked", trace.WithAttributes(
		attribute.String("media.key", key),
		attribute.Bool("media.exists", exists),
	))
	span.SetStatus(codes.Ok, "")

	return exists
}

func (fs *mediaFileSystem) UploadMedia(ctx context.Context, testId string, file entity.File) error {
	if err := assert.StringNotEmpty(testId); err != nil {
		return fmt.Errorf("testId must not be empty")
	}

	ctx, span := fs.tracer.Start(ctx, "mediaFileSystemRepository.UploadMedia")
	defer span.End()
	span.SetAttributes(
		attribute.String("test.id", testId),
		attribute.String("media.extension", file.GetFileExtension()),
	)

	// Upload to S3
	key := fs.getMediaKey(testId, file.GetFileExtension())
	if err := fs.s3Wrapper.UploadMediaFile(ctx, key, file.GetFileData(), nil); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to upload media file")
		return fmt.Errorf("failed to upload %s to S3: %w", file.GetFileExtension(), err)
	}

	span.AddEvent("media file uploaded", trace.WithAttributes(
		attribute.String("media.key", key),
	))
	span.SetStatus(codes.Ok, "")

	return nil
}
