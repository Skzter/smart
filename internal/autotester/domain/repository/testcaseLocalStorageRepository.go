package repository

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageRepository provides low-level file operations for managing TestCase files
// in local storage. The repository handles both data persistence (Save, Delete) and
// path provisioning for the test runner (GetTestPath*). All operations are user- and session-scoped.
type TestcaseLocalStorageRepository interface {
	// Save persists the given TestCase for the specified user and session.
	// Implementations may create directories or files as required.
	Save(testcase *entity.TestCase, userId, sessionId string) error

	// GetTestPath returns the validated relative file path to a specific TestCase file,
	// starting from the root of the local test storage directory.
	// The path is ready for the test runner to execute the test without
	// needing to know the internal storage structure.
	GetTestPath(testId, userId, sessionId string) (string, error)

	// GetTestPathsBySession returns all validated relative file paths to TestCase files for a session,
	// starting from the root of the local test storage directory.
	// The paths are ready for the test runner to execute all tests within the session
	// without needing to know the internal storage structure.
	GetTestPathsBySession(userId, sessionId string) ([]string, error)

	// GetTestPathsByUser returns all validated relative file paths to TestCase files for a user,
	// grouped by sessionId. The paths start from the root of the local test storage directory
	// and are ready for the test runner to execute all tests across sessions
	// without needing to know the internal storage structure.
	GetTestPathsByUser(userId string) (map[string][]string, error)

	// Delete removes the TestCase with the given testId for the provided
	// user and session. Implementations should return nil when the file did
	// not exist (idempotent delete) or an error for IO failures.
	Delete(testId, userId, sessionId string) error

	// DeleteOlderThan removes all TestCases that are older than the specified duration.
	// Returns the number of deleted files and any error that occurred.
	DeleteOlderThan(maxAge time.Duration) (int, error)
}

const testcaseLanguageDefault = "ts"

type testcaseLocalStorageRepository struct {
	filesystem FileSystem
	logger     *slog.Logger
}

// NewTestcaseLocalStorageRepository creates a new local storage repository instance for the testcase entity.
// Returns the repository or an error.
func NewTestcaseLocalStorageRepository(logger *slog.Logger, filesystem FileSystem) (TestcaseLocalStorageRepository, error) {
	if err := assert.NotNil(logger, filesystem); err != nil {
		return nil, fmt.Errorf("invalid args: %w", err)
	}

	return &testcaseLocalStorageRepository{
		filesystem: filesystem,
		logger:     logger,
	}, nil
}

func (r *testcaseLocalStorageRepository) Save(testcase *entity.TestCase, userId, sessionId string) error {
	if err := validatePathNameElements(userId, sessionId); err != nil {
		return fmt.Errorf("validate path elements: %w", err)
	}
	if err := validateTestcase(testcase); err != nil {
		return fmt.Errorf("validate testcase: %w", err)
	}

	filename := testcase.TestID + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid testcase filename %s: %w", filename, err)
	}

	dir := filepath.Join(userId, sessionId)
	if err := r.filesystem.MkdirAll(dir); err != nil {
		r.logger.Error("create directory failed", "dir", dir, "err", err)
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, filename)
	data := []byte(testcase.TestCode.Code)
	if err := r.filesystem.WriteFile(filePath, data); err != nil {
		r.logger.Error("write testcase file failed", "path", filePath, "err", err)
		return fmt.Errorf("write testcase file %s: %w", filePath, err)
	}

	return nil
}

func (r *testcaseLocalStorageRepository) GetTestPath(testId, userId, sessionId string) (string, error) {
	if err := validatePathNameElements(userId, sessionId); err != nil {
		return "", fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filename := testId + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		return "", fmt.Errorf("invalid testcase filename %s: %w", filename, err)
	}

	relativePath := filepath.Join(dir, filename)

	if _, err := r.filesystem.GetFileStats(relativePath); err != nil {
		r.logger.Error("testcase file not found", "path", relativePath, "err", err)
		return "", fmt.Errorf("testcase file not found %s: %w", relativePath, err)
	}

	validatedPath, err := r.filesystem.GetValidatedPath(relativePath)
	if err != nil {
		return "", fmt.Errorf("validate path: %w", err)
	}

	return validatedPath, nil
}

func (r *testcaseLocalStorageRepository) GetTestPathsBySession(userId, sessionId string) ([]string, error) {
	if err := validatePathNameElements(userId, sessionId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filenames, err := r.filesystem.ReadDir(dir)
	if err != nil {
		r.logger.Error("read directory failed", "dir", dir, "err", err)
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	results := make([]string, 0, len(filenames))
	for _, filename := range filenames {
		if err := validateFilename(filename); err != nil {
			r.logger.Warn("skipping invalid testcase file", "file", filename, "err", err)
			continue
		}

		relativePath := filepath.Join(dir, filename)
		validatedPath, err := r.filesystem.GetValidatedPath(relativePath)
		if err != nil {
			r.logger.Error("failed to validate path", "path", relativePath, "err", err)
			return nil, fmt.Errorf("validate path for %s: %w", relativePath, err)
		}

		results = append(results, validatedPath)
	}

	return results, nil
}

func (r *testcaseLocalStorageRepository) GetTestPathsByUser(userId string) (map[string][]string, error) {
	if err := validatePathNameElements(userId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	sessions, err := r.filesystem.ReadDir(userId)
	if err != nil {
		r.logger.Error("read user directory failed", "dir", userId, "err", err)
		return nil, fmt.Errorf("read user directory %s: %w", userId, err)
	}

	results := make(map[string][]string)

	for _, sessionId := range sessions {
		testPaths, err := r.GetTestPathsBySession(userId, sessionId)
		if err != nil {
			return nil, fmt.Errorf("read session %s for user %s: %w", sessionId, userId, err)
		}

		results[sessionId] = testPaths
	}
	return results, nil
}

func (r *testcaseLocalStorageRepository) Delete(testId, userId, sessionId string) error {
	if err := validatePathNameElements(testId, userId, sessionId); err != nil {
		return fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filename := testId + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid testcase filename %s: %w", filename, err)
	}

	path := filepath.Join(dir, filename)
	if err := r.filesystem.Remove(path, false); err != nil {
		r.logger.Error("remove file failed", "path", path, "err", err)
		return fmt.Errorf("remove file %s: %w", path, err)
	}
	return nil
}

func (r *testcaseLocalStorageRepository) DeleteOlderThan(maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)
	deletedCount := 0

	userIds, err := r.filesystem.ReadDir(".")
	if err != nil {
		return 0, fmt.Errorf("read user directories: %w", err)
	}

	for _, userId := range userIds {
		sessions, err := r.filesystem.ReadDir(userId)
		if err != nil {
			continue
		}

		for _, sessionId := range sessions {
			sessionPath := filepath.Join(userId, sessionId)

			files, err := r.filesystem.ReadDir(sessionPath)
			if err != nil {
				continue
			}

			for _, filename := range files {
				filePath := filepath.Join(sessionPath, filename)

				fileInfo, err := r.filesystem.GetFileStats(filePath)
				if err != nil {
					continue
				}

				if fileInfo.ModTime().Before(cutoffTime) {
					if err := r.filesystem.Remove(filePath, false); err != nil {
						continue
					}

					deletedCount++
				}
			}
		}
	}
	return deletedCount, nil
}

func validateTestcase(testcase *entity.TestCase) error {
	if err := assert.NotNil(testcase); err != nil {
		return fmt.Errorf("obj must not be nil: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestID); err != nil {
		return fmt.Errorf("testcase.TestID must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(testcase.TestCode.Code); err != nil {
		return fmt.Errorf("testcase.TestCode.Code must not be empty: %w", err)
	}
	if err := assert.StringNotEmpty(testcaseLanguageDefault); err != nil {
		return fmt.Errorf("testcase.TestCode.Language must not be empty: %w", err)
	}
	return nil
}

func validatePathNameElements(values ...string) error {
	for i, v := range values {
		if err := assert.StringNotEmpty(v); err != nil {
			return fmt.Errorf("path element %d must not be empty: %w", i, err)
		}
	}
	return nil
}

// validateFilename parses and validates a filename in the form
// <testId>.<language>. On success it returns the parsed
// testId and language.
func validateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("empty filename")
	}

	dotPosition := strings.LastIndex(filename, ".")
	if dotPosition <= 0 || dotPosition == len(filename)-1 {
		return fmt.Errorf("missing or invalid extension")
	}

	testId := filename[:dotPosition]
	if _, err := uuid.Parse(testId); err != nil {
		return fmt.Errorf("invalid UUID format in testId '%s': %w", testId, err)
	}

	return nil
}
