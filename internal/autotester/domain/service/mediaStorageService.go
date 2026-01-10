package service

import (
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// MediaStorageService provides business operations for managing media files
// (screenshots and videos) from Playwright test runs.
type MediaStorageService interface {
	// GetScreenshot returns the screenshot data for the given testId.
	// Returns an error if no screenshot exists or reading fails.
	GetScreenshot(testId string) ([]byte, error)

	// GetVideo returns the video data for the given testId.
	// Returns an error if no video exists or reading fails.
	GetVideo(testId string) ([]byte, error)

	// GetScreenshotPath returns the file path to the screenshot for the given testId.
	// Returns an error if no screenshot exists.
	GetScreenshotPath(testId string) (string, error)

	// GetVideoPath returns the file path to the video for the given testId.
	// Returns an error if no video exists.
	GetVideoPath(testId string) (string, error)

	// HasMedia checks if media files exist for the given testId.
	// Returns hasScreenshot and hasVideo flags.
	HasMedia(testId string) (hasScreenshot, hasVideo bool)

	// GetMediaDir returns the media directory path for the given testId.
	// Creates the directory if it does not exist.
	GetMediaDir(testId string) (string, error)

	// CleanupMedia removes all media files for the given testId.
	// Called before a new test run to clean up old failure media.
	CleanupMedia(testId string) error
}

type mediaStorageService struct {
	logger *slog.Logger
	repo   repository.MediaFileSystem
}

// NewMediaStorageService creates a new MediaStorageService using the provided
// logger and repository. Returns an error if required dependencies are nil.
func NewMediaStorageService(logger *slog.Logger, repo repository.MediaFileSystem) (MediaStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &mediaStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

func (s *mediaStorageService) GetScreenshot(testId string) ([]byte, error) {
	s.logger.Debug("getting screenshot",
		slog.String("testId", testId),
	)

	data, err := s.repo.GetScreenshot(testId)
	if err != nil {
		s.logger.Debug("failed to get screenshot",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get screenshot failed: %w", err)
	}

	s.logger.Debug("screenshot retrieved successfully",
		slog.String("testId", testId),
		slog.Int("size", len(data)),
	)
	return data, nil
}

func (s *mediaStorageService) GetVideo(testId string) ([]byte, error) {
	s.logger.Debug("getting video",
		slog.String("testId", testId),
	)

	data, err := s.repo.GetVideo(testId)
	if err != nil {
		s.logger.Debug("failed to get video",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get video failed: %w", err)
	}

	s.logger.Debug("video retrieved successfully",
		slog.String("testId", testId),
		slog.Int("size", len(data)),
	)
	return data, nil
}

func (s *mediaStorageService) GetScreenshotPath(testId string) (string, error) {
	s.logger.Debug("getting screenshot path",
		slog.String("testId", testId),
	)

	path, err := s.repo.GetScreenshotPath(testId)
	if err != nil {
		s.logger.Debug("failed to get screenshot path",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("get screenshot path failed: %w", err)
	}

	s.logger.Debug("screenshot path retrieved successfully",
		slog.String("testId", testId),
		slog.String("path", path),
	)
	return path, nil
}

func (s *mediaStorageService) GetVideoPath(testId string) (string, error) {
	s.logger.Debug("getting video path",
		slog.String("testId", testId),
	)

	path, err := s.repo.GetVideoPath(testId)
	if err != nil {
		s.logger.Debug("failed to get video path",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("get video path failed: %w", err)
	}

	s.logger.Debug("video path retrieved successfully",
		slog.String("testId", testId),
		slog.String("path", path),
	)
	return path, nil
}

func (s *mediaStorageService) HasMedia(testId string) (hasScreenshot, hasVideo bool) {
	hasScreenshot = s.repo.HasScreenshot(testId)
	hasVideo = s.repo.HasVideo(testId)

	s.logger.Debug("checked media availability",
		slog.String("testId", testId),
		slog.Bool("hasScreenshot", hasScreenshot),
		slog.Bool("hasVideo", hasVideo),
	)
	return hasScreenshot, hasVideo
}

func (s *mediaStorageService) GetMediaDir(testId string) (string, error) {
	s.logger.Debug("getting media directory",
		slog.String("testId", testId),
	)

	path, err := s.repo.GetMediaDir(testId)
	if err != nil {
		s.logger.Error("failed to get media directory",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("get media directory failed: %w", err)
	}

	s.logger.Debug("media directory retrieved successfully",
		slog.String("testId", testId),
		slog.String("path", path),
	)
	return path, nil
}

func (s *mediaStorageService) CleanupMedia(testId string) error {
	s.logger.Debug("cleaning up media",
		slog.String("testId", testId),
	)

	if err := s.repo.DeleteMedia(testId); err != nil {
		s.logger.Error("failed to cleanup media",
			slog.String("testId", testId),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("cleanup media failed: %w", err)
	}

	s.logger.Debug("media cleaned up successfully",
		slog.String("testId", testId),
	)
	return nil
}
