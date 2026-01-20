package repository

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageRepository provides low-level file operations for managing TestCase files
// in local storage. The repository handles both data persistence (Save, Delete) and
// path provisioning for the test runner (GetTestPath*). All operations are user- and chat-scoped.
type TestcaseLocalStorageRepository interface {
	// Save persists the given TestCase for the specified user and chat.
	// Implementations may create directories or files as required.
	Save(testcase *entity.TestCase, userId, chatId string) error

	// Read retrieves the content of a TestCase file from local storage.
	// The file is uniquely identified and validated using testId, userId, and chatId.
	// Returns the file content as a byte slice, or an error if the file does not exist or parameters are invalid.
	Read(testId, userId, chatId string) ([]byte, error)

	// GetTestPath returns the validated relative file path to a specific TestCase file,
	// starting from the root of the local test storage directory.
	// The path is ready for the test runner to execute the test without
	// needing to know the internal storage structure.
	GetTestPath(testId, userId, chatId string) (string, error)

	// GetTestPathsByChat returns all validated relative file paths to TestCase files for a chat,
	// starting from the root of the local test storage directory.
	// The paths are ready for the test runner to execute all tests within the chat
	// without needing to know the internal storage structure.
	GetTestPathsByChat(userId, chatId string) ([]string, error)

	// GetTestPathsByUser returns all validated relative file paths to TestCase files for a user,
	// grouped by chatId. The paths start from the root of the local test storage directory
	// and are ready for the test runner to execute all tests across chats
	// without needing to know the internal storage structure.
	GetTestPathsByUser(userId string) (map[string][]string, error)

	// Delete removes the TestCase with the given testId for the provided
	// user and chat. Implementations should return nil when the file did
	// not exist (idempotent delete) or an error for IO failures.
	Delete(testId, userId, chatId string) error

	// DeleteOlderThan removes all TestCases that are older than the specified duration.
	// Returns the number of deleted files and any error that occurred.
	DeleteOlderThan(maxAge time.Duration) (int, error)
}

const testcaseLanguageDefault = "spec.ts"

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

func (r *testcaseLocalStorageRepository) Save(testcase *entity.TestCase, userId, chatId string) error {
	if err := validatePathNameElements(userId, chatId); err != nil {
		return fmt.Errorf("validate path elements: %w", err)
	}
	if err := validateTestcase(testcase); err != nil {
		r.logger.Error(fmt.Sprintf("validate testcase: %s", err))
		return errors.ErrValidation
	}

	filename := testcase.TestID + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		r.logger.Error(fmt.Sprintf("invalid testcase filename %s: %v", filename, err))
		return errors.ErrValidation
	}

	dir := filepath.Join(userId, chatId)
	if err := r.filesystem.MkdirAll(dir); err != nil {
		r.logger.Error("create directory failed", "dir", dir, "err", err)
		return errors.ErrInternalServer
	}

	filePath := filepath.Join(dir, filename)
	data := []byte(testcase.TestCode.Code)
	if err := r.filesystem.WriteFile(filePath, data); err != nil {
		r.logger.Error("write testcase file failed", "path", filePath, "err", err)
		return errors.ErrInternalServer
	}

	return nil
}

func (r *testcaseLocalStorageRepository) Read(testId, userId, chatId string) ([]byte, error) {
	if err := validatePathNameElements(userId, chatId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, chatId)
	filename := testId + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		r.logger.Error(fmt.Sprintf("invalid testcase filename %s: %s", filename, err))
		return nil, errors.ErrValidation
	}

	relativePath := filepath.Join(dir, filename)

	fileContent, err := r.filesystem.ReadFile(relativePath)
	if err != nil {
		r.logger.Error(fmt.Sprintf("read file failed, path: %s", relativePath))
		return nil, errors.ErrInternalServer
	}

	return fileContent, nil
}

func (r *testcaseLocalStorageRepository) GetTestPath(testId, userId, chatId string) (string, error) {
	if err := validatePathNameElements(userId, chatId); err != nil {
		return "", fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, chatId)
	filename := testId + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		r.logger.Error(fmt.Sprintf("invalid testcase filename: %s %v", filename, err))
		return "", errors.ErrValidation
	}

	relativePath := filepath.Join(dir, filename)

	if _, err := r.filesystem.GetFileStats(relativePath); err != nil {
		r.logger.Error(fmt.Sprintf("testcase file not found: %s %v", relativePath, err))
		return "", errors.ErrValidation
	}

	validatedPath, err := r.filesystem.GetValidatedPath(relativePath)
	if err != nil {
		r.logger.Error(fmt.Sprintf("validate path: %s", err))
		return "", errors.ErrInternalServer
	}

	return validatedPath, nil
}

func (r *testcaseLocalStorageRepository) GetTestPathsByChat(userId, chatId string) ([]string, error) {
	if err := validatePathNameElements(userId, chatId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, chatId)
	filenames, err := r.filesystem.ReadDir(dir)
	if err != nil {
		r.logger.Error(fmt.Sprintf("read directory failed: %s %v", dir, err))
		return nil, errors.ErrInternalServer
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
			return nil, errors.ErrInternalServer
		}

		results = append(results, validatedPath)
	}

	return results, nil
}

func (r *testcaseLocalStorageRepository) GetTestPathsByUser(userId string) (map[string][]string, error) {
	if err := validatePathNameElements(userId); err != nil {
		r.logger.Error(fmt.Sprintf("validate path elements: %s", err))
		return nil, errors.ErrValidation
	}

	chats, err := r.filesystem.ReadDir(userId)
	if err != nil {
		r.logger.Error("read user directory failed", "dir", userId, "err", err)
		return nil, errors.ErrInternalServer
	}

	results := make(map[string][]string)

	for _, chatId := range chats {
		testPaths, err := r.GetTestPathsByChat(userId, chatId)
		if err != nil {
			return nil, fmt.Errorf("read chat %s for user %s: %w", chatId, userId, err)
		}

		results[chatId] = testPaths
	}
	return results, nil
}

func (r *testcaseLocalStorageRepository) Delete(testId, userId, chatId string) error {
	if err := validatePathNameElements(testId, userId, chatId); err != nil {
		return fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, chatId)
	filename := testId + "." + testcaseLanguageDefault
	if err := validateFilename(filename); err != nil {
		r.logger.Error(fmt.Sprintf("invalid testcase filename %s: %s", filename, err))
		return errors.ErrValidation
	}

	path := filepath.Join(dir, filename)
	if err := r.filesystem.Remove(path, false); err != nil {
		r.logger.Error("remove file failed", "path", path, "err", err)
		return errors.ErrInternalServer
	}
	return nil
}

func (r *testcaseLocalStorageRepository) DeleteOlderThan(maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)
	deletedCount := 0

	userIds, err := r.filesystem.ReadDir(".")
	if err != nil {
		r.logger.Error(fmt.Sprintf("read user directories: %s", err))
		return 0, errors.ErrInternalServer
	}

	for _, userId := range userIds {
		chats, err := r.filesystem.ReadDir(userId)
		if err != nil {
			r.logger.Warn("failed to read user directory", "userId", userId, "err", err)
			continue
		}

		for _, chatId := range chats {
			chatPath := filepath.Join(userId, chatId)
			files, err := r.filesystem.ReadDir(chatPath)
			if err != nil {
				r.logger.Warn("failed to read chat directory", "chatPath", chatPath, "err", err)
				continue
			}

			for _, filename := range files {
				filePath := filepath.Join(chatPath, filename)
				fileInfo, err := r.filesystem.GetFileStats(filePath)
				if err != nil || !fileInfo.ModTime().Before(cutoffTime) {
					continue
				}

				if err := r.filesystem.Remove(filePath, false); err != nil {
					r.logger.Warn("failed to delete old file", "path", filePath, "err", err)
					continue
				}

				deletedCount++
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

	dotPosition := len(filename) - len(testcaseLanguageDefault) - 1
	if !strings.HasSuffix(filename, testcaseLanguageDefault) {
		return fmt.Errorf("missing or invalid extension")
	}

	testId := filename[:dotPosition]
	if _, err := uuid.Parse(testId); err != nil {
		return fmt.Errorf("invalid UUID format in testId '%s': %w", testId, err)
	}

	return nil
}
