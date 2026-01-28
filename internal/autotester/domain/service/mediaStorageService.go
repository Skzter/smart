package service

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// MediaStorageService provides business operations for managing media files
// (screenshots and videos) from Playwright test runs stored in S3.
type MediaStorageService interface {
	// GetScreenshotUrl returns the S3 URL to the screenshot for the given testId.
	// Returns an error if no screenshot exists.
	GetScreenshotUrl(ctx context.Context, testId string) (string, error)

	// GetVideoUrl returns the S3 URL to the video for the given testId.
	// Returns an error if no video exists.
	GetVideoUrl(ctx context.Context, testId string) (string, error)

	// HasMedia checks if media files exist for the given testId in S3.
	// Returns hasScreenshot and hasVideo flags.
	HasMedia(ctx context.Context, testId string) (hasScreenshot, hasVideo bool)

	// UploadMedia uploads a mediafile
	UploadMedia(ctx context.Context, testId string, file entity.File) error
}

type mediaStorageService struct {
	logger *slog.Logger
	repo   repository.MediaFileSystem
	tracer trace.Tracer
}

// NewMediaStorageService creates a new MediaStorageService using the provided
// logger and repository. Returns an error if required dependencies are nil.
func NewMediaStorageService(logger *slog.Logger, repo repository.MediaFileSystem, tracer trace.Tracer) (MediaStorageService, error) {
	if err := assert.NotNil(logger, repo, tracer); err != nil {
		return nil, err
	}

	return &mediaStorageService{
		logger: logger,
		repo:   repo,
		tracer: tracer,
	}, nil
}

func (s *mediaStorageService) GetScreenshotUrl(ctx context.Context, testId string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "mediaStorageService.GetScreenshotUrl")
	defer span.End()

	s.logger.Debug("getting screenshot URL",
		slog.String("testId", testId),
	)

	url, err := s.repo.GetScreenshotUrl(ctx, testId)
	if err != nil {
		s.logger.Debug("failed to get screenshot URL",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "error getting screenshot URL")
		return "", fmt.Errorf("get screenshot URL failed: %w", err)
	}

	s.logger.Debug("screenshot URL retrieved successfully",
		slog.String("testId", testId),
		slog.String("url", url),
	)
	span.SetStatus(codes.Ok, "")
	return url, nil
}

func (s *mediaStorageService) GetVideoUrl(ctx context.Context, testId string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "mediaStorageService.GetVideoUrl")
	defer span.End()

	s.logger.Debug("getting video URL",
		slog.String("testId", testId),
	)

	url, err := s.repo.GetVideoUrl(ctx, testId)
	if err != nil {
		s.logger.Debug("failed to get video URL",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "error getting video URL")
		return "", fmt.Errorf("get video URL failed: %w", err)
	}

	s.logger.Debug("video URL retrieved successfully",
		slog.String("testId", testId),
		slog.String("url", url),
	)
	span.SetStatus(codes.Ok, "")
	return url, nil
}

func (s *mediaStorageService) HasMedia(ctx context.Context, testId string) (hasScreenshot, hasVideo bool) {
	ctx, span := s.tracer.Start(ctx, "mediaStorageService.HasMedia")
	defer span.End()

	hasScreenshot = s.repo.HasScreenshot(ctx, testId)
	hasVideo = s.repo.HasVideo(ctx, testId)

	s.logger.Debug("checked media availability",
		slog.String("testId", testId),
		slog.Bool("hasScreenshot", hasScreenshot),
		slog.Bool("hasVideo", hasVideo),
	)
	span.SetStatus(codes.Ok, "")
	return hasScreenshot, hasVideo
}

func (s *mediaStorageService) UploadMedia(ctx context.Context, testId string, file entity.File) error {
	ctx, span := s.tracer.Start(ctx, "mediaStorageService.UploadMedia")
	defer span.End()

	s.logger.Debug("uploading media file to S3",
		slog.String("testId", testId),
	)

	err := s.repo.UploadMedia(ctx, testId, file)
	if err != nil {
		s.logger.Error("failed to upload media files",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "error uploading media")
		return fmt.Errorf("upload media failed: %w", err)
	}

	s.logger.Debug("media files uploaded successfully",
		slog.String("testId", testId),
	)
	span.SetStatus(codes.Ok, "")
	return nil
}
