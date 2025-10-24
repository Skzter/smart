package repository

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageRepository defines methods for saving, reading and deleting
// local TestCase objects. The Save method returns the generated local key on success.
type TestcaseLocalStorageRepository interface {
	// Save persists the given TestCase for the specified user and session.
	// Implementations may create directories or files as required.
	Save(testcase *entity.TestCase, userId, sessionId string) error

	// ReadLatest returns the most recent TestCase matching testId for the
	// provided user and session. Returns an error when not found or on IO
	// failure.
	Read(testId, lang, userId, sessionId string) (*entity.TestCase, error)

	// ReadAllBySession returns all TestCases for the given user within the
	// specified session. The returned slice may be empty if no entries exist.
	ReadAllBySession(userId, sessionId string) ([]*entity.TestCase, error)

	// ReadAllByUser returns all TestCases for the given user across all
	// sessions. The returned slice may be empty when none exist.
	ReadAllByUser(userId string) (map[string][]*entity.TestCase, error)

	// Delete removes the TestCase with the given testId for the provided
	// user and session. Implementations should return nil when the file did
	// not exist (idempotent delete) or an error for IO failures.
	Delete(testId, lang, userId, sessionId string) error
}

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

	filename := testcase.TestID + "." + testcase.TestCode.Language
	if _, _, err := validateFilename(filename); err != nil {
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

func (r *testcaseLocalStorageRepository) Read(testId, lang, userId, sessionId string) (*entity.TestCase, error) {
	if err := validatePathNameElements(userId, sessionId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filename := testId + "." + lang
	if _, _, err := validateFilename(filename); err != nil {
		return nil, fmt.Errorf("invalid testcase filename %s: %w", filename, err)
	}

	path := filepath.Join(dir, filename)
	return r.readTestcaseFromPath(path, testId, lang)
}

func (r *testcaseLocalStorageRepository) ReadAllBySession(userId, sessionId string) ([]*entity.TestCase, error) {
	if err := validatePathNameElements(userId, sessionId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filenames, err := r.filesystem.ReadDir(dir)
	if err != nil {
		r.logger.Error("read directory failed", "dir", dir, "err", err)
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	results := make([]*entity.TestCase, 0, len(filenames))
	for _, filename := range filenames {
		testId, lang, err := validateFilename(filename)
		if err != nil {
			r.logger.Warn("skipping invalid testcase file", "file", filename, "err", err)
			continue
		}

		path := filepath.Join(dir, filename)
		tc, err := r.readTestcaseFromPath(path, testId, lang)
		if err != nil {
			r.logger.Error("failed to read testcase from path", "path", path, "err", err)
			return nil, fmt.Errorf("read testcase %s: %w", path, err)
		}

		results = append(results, tc)
	}

	return results, nil
}

func (r *testcaseLocalStorageRepository) ReadAllByUser(userId string) (map[string][]*entity.TestCase, error) {
	if err := validatePathNameElements(userId); err != nil {
		return nil, fmt.Errorf("validate path elements: %w", err)
	}

	sessions, err := r.filesystem.ReadDir(userId)
	if err != nil {
		r.logger.Error("read user directory failed", "dir", userId, "err", err)
		return nil, fmt.Errorf("read user directory %s: %w", userId, err)
	}

	results := make(map[string][]*entity.TestCase)

	for _, sessionId := range sessions {
		testcases, err := r.ReadAllBySession(userId, sessionId)
		if err != nil {
			return nil, fmt.Errorf("read session %s for user %s: %w", sessionId, userId, err)
		}

		results[sessionId] = testcases
	}
	return results, nil
}

func (r *testcaseLocalStorageRepository) Delete(testId, lang, userId, sessionId string) error {
	if err := validatePathNameElements(testId, userId, sessionId); err != nil {
		return fmt.Errorf("validate path elements: %w", err)
	}

	dir := filepath.Join(userId, sessionId)
	filename := testId + "." + lang
	if _, _, err := validateFilename(filename); err != nil {
		return fmt.Errorf("invalid testcase filename %s: %w", filename, err)
	}

	path := filepath.Join(dir, filename)
	if err := r.filesystem.Remove(path, false); err != nil {
		r.logger.Error("remove file failed", "path", path, "err", err)
		return fmt.Errorf("remove file %s: %w", path, err)
	}
	return nil
}

// readTestcaseFromPath reads the file at the given path and constructs a TestCase
// using the provided testId and language. The function assumes the caller has
// already validated the path/filename and therefore does not perform those checks.
func (r *testcaseLocalStorageRepository) readTestcaseFromPath(path, testId, lang string) (*entity.TestCase, error) {
	data, err := r.filesystem.ReadFile(path)
	if err != nil {
		r.logger.Error("read testcase file failed", "path", path, "err", err)
		return nil, fmt.Errorf("read testcase file %s: %w", path, err)
	}

	testcase := &entity.TestCase{
		TestID:      testId,
		Description: "",
		TestCode: entity.TestCode{
			Code:     string(data),
			Language: lang,
		},
		Status: entity.TestStatusNotRun,
	}

	return testcase, nil
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
	if err := assert.StringNotEmpty(testcase.TestCode.Language); err != nil {
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
func validateFilename(filename string) (testID, language string, err error) {
	if filename == "" {
		return "", "", fmt.Errorf("empty filename")
	}

	dotPosition := strings.LastIndex(filename, ".")
	if dotPosition <= 0 || dotPosition == len(filename)-1 {
		return "", "", fmt.Errorf("missing or invalid extension")
	}

	language = filename[dotPosition+1:]
	if language == "" {
		return "", "", fmt.Errorf("empty language")
	}

	testId := filename[:dotPosition]
	if _, err := uuid.Parse(testId); err != nil {
		return "", "", fmt.Errorf("invalid UUID format in testId '%s': %w", testId, err)
	}

	return testId, language, nil
}
