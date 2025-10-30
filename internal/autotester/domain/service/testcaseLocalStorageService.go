package service

import (
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageService provides business operations for managing TestCase files
// in local storage. The service handles both frontend API operations (Save, Delete) and
// test runner path provisioning (GetTestPath*). All operations are user- and session-scoped.
type TestcaseLocalStorageService interface {
	// Save stores the given TestCase for the specified user and session.
	// Returns an error when validation or storage fails.
	Save(testcase *entity.TestCase, userId, sessionId string) error

	// GetTestPath returns the relative file path to a specific TestCase file,
	// starting from the root of the local test storage directory.
	// Returns an error if path validation fails or the file does not exist.
	GetTestPath(testId, lang, userId, sessionId string) (string, error)

	// GetTestPathsBySession returns all relative TestCase file paths for a session,
	// starting from the root of the local test storage directory.
	// Returns an error if the session directory cannot be read or path validation fails.
	GetTestPathsBySession(userId, sessionId string) ([]string, error)

	// GetTestPathsByUser returns all relative TestCase file paths for a user, grouped by sessionId.
	// Paths start from the root of the local test storage directory.
	// Returns an error if the user directory cannot be read or any session directory read fails.
	GetTestPathsByUser(userId string) (map[string][]string, error)

	// Delete removes a testcase identified by testId for the given user and session.
	// Returns an error if path validation fails or the file cannot be deleted.
	Delete(testId, lang, userId, sessionId string) error

	// CleanupOldTests removes all testcases that are older than 24 hours.
	// Returns an error if the cleanup operation fails.
	CleanupOldTests() error
}

type testcaseLocalStorageService struct {
	logger *slog.Logger
	repo   repository.TestcaseLocalStorageRepository
}

// NewTestcaseLocalStorageService creates a TestcaseLocalStorageService using
// the provided logger and repository. It returns an error when required
// dependencies are nil.
func NewTestcaseLocalStorageService(logger *slog.Logger, repo repository.TestcaseLocalStorageRepository) (TestcaseLocalStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &testcaseLocalStorageService{
		logger: logger,
		repo:   repo,
	}, nil
}

func (s *testcaseLocalStorageService) Save(testcase *entity.TestCase, userId, sessionId string) error {
	s.logger.Debug("saving testcase",
		slog.String("testId", testcase.TestID),
		slog.String("language", testcase.TestCode.Language),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)

	if err := s.repo.Save(testcase, userId, sessionId); err != nil {
		s.logger.Error("failed to save testcase",
			slog.String("testId", testcase.TestID),
			slog.String("language", testcase.TestCode.Language),
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("save to repository failed: %w", err)
	}

	s.logger.Info("testcase saved successfully",
		slog.String("testId", testcase.TestID),
		slog.String("language", testcase.TestCode.Language),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)
	return nil
}

func (s *testcaseLocalStorageService) GetTestPath(testId, lang, userId, sessionId string) (string, error) {
	s.logger.Debug("getting testcase path",
		slog.String("testId", testId),
		slog.String("language", lang),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)

	path, err := s.repo.GetTestPath(testId, lang, userId, sessionId)
	if err != nil {
		s.logger.Error("failed to get testcase path",
			slog.String("testId", testId),
			slog.String("language", lang),
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return "", fmt.Errorf("get path from repository failed: %w", err)
	}

	s.logger.Debug("testcase path retrieved successfully",
		slog.String("testId", testId),
		slog.String("language", lang),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
		slog.String("path", path),
	)
	return path, nil
}

func (s *testcaseLocalStorageService) GetTestPathsBySession(userId, sessionId string) ([]string, error) {
	s.logger.Debug("getting all testcase paths for session",
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)

	paths, err := s.repo.GetTestPathsBySession(userId, sessionId)
	if err != nil {
		s.logger.Error("failed to get testcase paths by session",
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get paths from repository failed: %w", err)
	}

	s.logger.Info("testcase paths retrieved successfully",
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
		slog.Int("count", len(paths)),
	)
	return paths, nil
}

func (s *testcaseLocalStorageService) GetTestPathsByUser(userId string) (map[string][]string, error) {
	s.logger.Debug("getting all testcase paths for user",
		slog.String("userId", userId),
	)

	paths, err := s.repo.GetTestPathsByUser(userId)
	if err != nil {
		s.logger.Error("failed to get testcase paths by user",
			slog.String("userId", userId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("get paths from repository failed: %w", err)
	}

	totalCount := 0
	for _, sessionPaths := range paths {
		totalCount += len(sessionPaths)
	}

	s.logger.Info("testcase paths retrieved successfully",
		slog.String("userId", userId),
		slog.Int("sessionCount", len(paths)),
		slog.Int("totalPaths", totalCount),
	)
	return paths, nil
}

func (s *testcaseLocalStorageService) Delete(testId, lang, userId, sessionId string) error {
	if err := s.repo.Delete(testId, lang, userId, sessionId); err != nil {
		s.logger.Error("failed to delete testcase",
			slog.String("testId", testId),
			slog.String("language", lang),
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("delete from repository failed: %w", err)
	}

	s.logger.Info("testcase deleted successfully",
		slog.String("testId", testId),
		slog.String("language", lang),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)
	return nil
}

func (s *testcaseLocalStorageService) CleanupOldTests() error {
	const maxAge = 24 * time.Hour

	s.logger.Info("starting cleanup of old tests",
		slog.Duration("maxAge", maxAge),
	)

	deletedCount, err := s.repo.DeleteOlderThan(maxAge)
	if err != nil {
		s.logger.Error("cleanup of old tests failed",
			slog.Duration("maxAge", maxAge),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("cleanup old tests failed: %w", err)
	}

	s.logger.Info("cleanup of old tests completed successfully",
		slog.Int("deletedCount", deletedCount),
		slog.Duration("maxAge", maxAge),
	)
	return nil
}
