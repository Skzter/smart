package repository

import (
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// TestcaseLocalStorageRepository defines methods for saving, reading and deleting
// local TestCase objects. The Save method returns the generated local key on success.
type TestcaseLocalStorageRepository interface {
	// Save persists the given TestCase for the specified user and session.
	// Implementations may create directories or files as required.
	Save(testcase *entity.TestCase, userId string, sessionId string) error

	// ReadLatest returns the most recent TestCase matching testId for the
	// provided user and session. Returns an error when not found or on IO
	// failure.
	ReadLatest(testId string, userId string, sessionId string) (entity.TestCase, error)

	// ReadAllBySession returns all TestCases for the given user within the
	// specified session. The returned slice may be empty if no entries exist.
	ReadAllBySession(userId string, sessionId string) ([]entity.TestCase, error)

	// ReadAllByUser returns all TestCases for the given user across all
	// sessions. The returned slice may be empty when none exist.
	ReadAllByUser(userId string) ([]entity.TestCase, error)

	// Delete removes the TestCase with the given testId for the provided
	// user and session. Implementations should return nil when the file did
	// not exist (idempotent delete) or an error for IO failures.
	Delete(testId string, userId string, sessionId string) error

	// GetAll returns all TestCases for the given user and optional session.
	// If sessionId is empty, results across all sessions for the user are
	// returned.
	GetAll(userId string, sessionId string) ([]entity.TestCase, error)
}

type testcaseLocalStorageRepository struct {
	logger *slog.Logger
}

// NewTestcaseLocalStorageRepository creates a new local storage repository instance for the testcase entity.
// Returns the repository or an error.
func NewTestcaseLocalStorageRepository(logger *slog.Logger) (TestcaseLocalStorageRepository, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &testcaseLocalStorageRepository{
		logger: logger,
	}, nil
}

func (r *testcaseLocalStorageRepository) Save(testcase *entity.TestCase, userId string, sessionId string) error {
	panic("unimplmented")
}

func (r *testcaseLocalStorageRepository) ReadAllBySession(userId string, sessionId string) ([]entity.TestCase, error) {
	panic("unimplemented")
}

func (r *testcaseLocalStorageRepository) ReadAllByUser(userId string) ([]entity.TestCase, error) {
	panic("unimplemented")
}

func (r *testcaseLocalStorageRepository) ReadLatest(testId string, userId string, sessionId string) (entity.TestCase, error) {
	panic("unimplemented")
}

func (r *testcaseLocalStorageRepository) Delete(testId string, userId string, sessionId string) error {
	panic("unimplmented")
}

func (r *testcaseLocalStorageRepository) GetAll(userId string, sessionId string) ([]entity.TestCase, error) {
	panic("unimplmented")
}
