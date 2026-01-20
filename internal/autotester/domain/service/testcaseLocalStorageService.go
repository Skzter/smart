package service

import (
	"log/slog"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageService provides business operations for managing TestCase files
// in local storage. The service handles both frontend API operations (Save, Delete) and
// test runner path provisioning (GetTestPath*). All operations are user- and chat-scoped.
type TestcaseLocalStorageService interface {
	// Save stores the given TestCase for the specified user and chat.
	// Returns an error when validation or storage fails.
	Save(testcase *entity.TestCase, userId, chatId string) error

	// Read loads the content of a specific testcase identified by testId for the given user and chat.
	// Returns the file content as a string, or an error if reading or validation fails.
	Read(testId, userId, chatId string) (string, error)

	// GetTestPath returns the relative file path to a specific TestCase file,
	// starting from the root of the local test storage directory.
	// Returns an error if path validation fails or the file does not exist.
	GetTestPath(testId, userId, chatId string) (string, error)

	// GetTestPathsBychat returns all relative TestCase file paths for a chat,
	// starting from the root of the local test storage directory.
	// Returns an error if the Chat directory cannot be read or path validation fails.
	GetTestPathsByChat(userId, chatId string) ([]string, error)

	// GetTestPathsByUser returns all relative TestCase file paths for a user, grouped by chatId.
	// Paths start from the root of the local test storage directory.
	// Returns an error if the user directory cannot be read or any chat directory read fails.
	GetTestPathsByUser(userId string) (map[string][]string, error)

	// Delete removes a testcase identified by testId for the given user and chat.
	// Returns an error if path validation fails or the file cannot be deleted.
	Delete(testId, userId, chatId string) error

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
func NewTestcaseLocalStorageService(logger *slog.Logger, repo repository.TestcaseLocalStorageRepository, enableCleanup bool) (TestcaseLocalStorageService, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	service := &testcaseLocalStorageService{
		logger: logger,
		repo:   repo,
	}

	if enableCleanup {
		ticker := time.NewTicker(time.Hour * 24)
		go func() {
			for range ticker.C {
				if err := service.CleanupOldTests(); err != nil {
					service.logger.Error("failed to cleanup old tests",
						slog.String("error", err.Error()),
					)
				}
			}
		}()
	}

	return service, nil
}

func (s *testcaseLocalStorageService) Save(testcase *entity.TestCase, userId, chatId string) error {
	s.logger.Debug("saving testcase",
		slog.String("testId", testcase.TestID),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)

	if err := s.repo.Save(testcase, userId, chatId); err != nil {
		s.logger.Error("failed to save testcase",
			slog.String("testId", testcase.TestID),
			slog.String("userId", userId),
			slog.String("chatId", chatId),
			slog.String("error", err.Error()),
		)
		return errors.ErrInternalServer
	}

	s.logger.Debug("testcase saved successfully",
		slog.String("testId", testcase.TestID),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)
	return nil
}

func (s *testcaseLocalStorageService) Read(testId, userId, chatId string) (string, error) {
	s.logger.Debug("reading testcase",
		slog.String("testId", testId),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)

	fileContent, err := s.repo.Read(testId, userId, chatId)
	if err != nil {
		s.logger.Error("failed to read testcase file",
			slog.String("testId", testId),
			slog.String("userId", userId),
			slog.String("chatId", chatId),
			slog.String("error", err.Error()),
		)
		return "", errors.ErrInternalServer
	}

	s.logger.Debug("testcase file read successfully",
		slog.String("testId", testId),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
		slog.String("path", string(fileContent)),
	)
	return string(fileContent), nil
}

func (s *testcaseLocalStorageService) GetTestPath(testId, userId, chatId string) (string, error) {
	s.logger.Debug("getting testcase path",
		slog.String("testId", testId),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)

	path, err := s.repo.GetTestPath(testId, userId, chatId)
	if err != nil {
		s.logger.Error("failed to get testcase path",
			slog.String("testId", testId),
			slog.String("userId", userId),
			slog.String("chatId", chatId),
			slog.String("error", err.Error()),
		)
		return "", errors.ErrInternalServer
	}

	s.logger.Debug("testcase path retrieved successfully",
		slog.String("testId", testId),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
		slog.String("path", path),
	)
	return path, nil
}

func (s *testcaseLocalStorageService) GetTestPathsByChat(userId, chatId string) ([]string, error) {
	s.logger.Debug("getting all testcase paths for chat",
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)

	paths, err := s.repo.GetTestPathsByChat(userId, chatId)
	if err != nil {
		s.logger.Error("failed to get testcase paths by chat",
			slog.String("userId", userId),
			slog.String("chatId", chatId),
			slog.String("error", err.Error()),
		)
		return nil, errors.ErrInternalServer
	}

	s.logger.Debug("testcase paths retrieved successfully",
		slog.String("userId", userId),
		slog.String("chatId", chatId),
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
		return nil, errors.ErrInternalServer
	}

	totalCount := 0
	for _, chatPaths := range paths {
		totalCount += len(chatPaths)
	}

	s.logger.Debug("testcase paths retrieved successfully",
		slog.String("userId", userId),
		slog.Int("chatCount", len(paths)),
		slog.Int("totalPaths", totalCount),
	)
	return paths, nil
}

func (s *testcaseLocalStorageService) Delete(testId, userId, chatId string) error {
	if err := s.repo.Delete(testId, userId, chatId); err != nil {
		s.logger.Error("failed to delete testcase",
			slog.String("testId", testId),
			slog.String("userId", userId),
			slog.String("chatId", chatId),
			slog.String("error", err.Error()),
		)
		return errors.ErrInternalServer
	}

	s.logger.Debug("testcase deleted successfully",
		slog.String("testId", testId),
		slog.String("userId", userId),
		slog.String("chatId", chatId),
	)
	return nil
}

func (s *testcaseLocalStorageService) CleanupOldTests() error {
	const maxAge = 24 * time.Hour

	s.logger.Debug("starting cleanup of old tests",
		slog.Duration("maxAge", maxAge),
	)

	deletedCount, err := s.repo.DeleteOlderThan(maxAge)
	if err != nil {
		s.logger.Error("cleanup of old tests failed",
			slog.Duration("maxAge", maxAge),
			slog.String("error", err.Error()),
		)
		return errors.ErrInternalServer
	}

	s.logger.Debug("cleanup of old tests completed successfully",
		slog.Int("deletedCount", deletedCount),
		slog.Duration("maxAge", maxAge),
	)
	return nil
}
