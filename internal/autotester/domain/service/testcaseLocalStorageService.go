package service

import (
	"log/slog"

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
	return s.repo.Save(testcase, userId, sessionId)
}

func (s *testcaseLocalStorageService) Read(testId, lang, userId, sessionId string) (*entity.TestCase, error) {
	return s.repo.Read(testId, lang, userId, sessionId)
}

func (s *testcaseLocalStorageService) ReadAllBySession(userId, sessionId string) ([]*entity.TestCase, error) {
	return s.repo.ReadAllBySession(userId, sessionId)
}

func (s *testcaseLocalStorageService) ReadAllByUser(userId string) (map[string][]*entity.TestCase, error) {
	return s.repo.ReadAllByUser(userId)
}

func (s *testcaseLocalStorageService) Delete(testId, lang, userId, sessionId string) error {
	return s.repo.Delete(testId, lang, userId, sessionId)
}
