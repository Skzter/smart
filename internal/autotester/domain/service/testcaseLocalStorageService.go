package service

import (
	"fmt"
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageService defines business operations for storing and
// reading TestCase objects in a local storage backend. The service is
// user- and session-scoped and exposes simple read/write entry points.
type TestcaseLocalStorageService interface {
	// Save stores the given TestCase for the provided user and session.
	// Returns an error when validation or storage fails.
	Save(testcase *entity.TestCase, userId, sessionId string) error

	// ReadLatest returns the latest TestCase for the given testId within the
	// specified user and session. The returned TestCase is the most recent
	// one for that test identifier.
	Read(testId, lang, userId, sessionId string) (*entity.TestCase, error)

	// ReadAllBySession returns all TestCases that belong to the given user
	// and session. The slice may be empty when no testcases are found.
	ReadAllBySession(userId, sessionId string) ([]*entity.TestCase, error)

	// ReadAllByUser returns all TestCases that belong to the given user
	// across all sessions. The slice may be empty when no testcases exist.
	ReadAllByUser(userId string) (map[string][]*entity.TestCase, error)

	// Delete removes a testcase identified by testId for the given user and session.
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

func (s *testcaseLocalStorageService) Read(testId, lang, userId, sessionId string) (*entity.TestCase, error) {
	s.logger.Debug("reading testcase",
		slog.String("testId", testId),
		slog.String("language", lang),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)

	testcase, err := s.repo.Read(testId, lang, userId, sessionId)
	if err != nil {
		s.logger.Error("failed to read testcase",
			slog.String("testId", testId),
			slog.String("language", lang),
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("read from repository failed: %w", err)
	}

	s.logger.Debug("testcase read successfully",
		slog.String("testId", testId),
		slog.String("language", lang),
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)
	return testcase, nil
}

func (s *testcaseLocalStorageService) ReadAllBySession(userId, sessionId string) ([]*entity.TestCase, error) {
	s.logger.Debug("reading all testcases for session",
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
	)

	testcases, err := s.repo.ReadAllBySession(userId, sessionId)
	if err != nil {
		s.logger.Error("failed to read testcases by session",
			slog.String("userId", userId),
			slog.String("sessionId", sessionId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("read from repository failed: %w", err)
	}

	s.logger.Info("testcases read successfully",
		slog.String("userId", userId),
		slog.String("sessionId", sessionId),
		slog.Int("count", len(testcases)),
	)
	return testcases, nil
}

func (s *testcaseLocalStorageService) ReadAllByUser(userId string) (map[string][]*entity.TestCase, error) {
	s.logger.Debug("reading all testcases for user",
		slog.String("userId", userId),
	)

	testcases, err := s.repo.ReadAllByUser(userId)
	if err != nil {
		s.logger.Error("failed to read testcases by user",
			slog.String("userId", userId),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("read from repository failed: %w", err)
	}

	totalCount := 0
	for _, sessionTestcases := range testcases {
		totalCount += len(sessionTestcases)
	}

	s.logger.Info("testcases read successfully",
		slog.String("userId", userId),
		slog.Int("sessionCount", len(testcases)),
		slog.Int("totalTestcases", totalCount),
	)
	return testcases, nil
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
