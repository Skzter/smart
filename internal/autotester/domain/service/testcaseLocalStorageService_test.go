package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

// nolint:dupl
func TestNewTestcaseLocalStorageService(t *testing.T) {
	logger := slog.Default()
	mockRepo := &mocks.MockTestcaseLocalStorageRepository{}

	tests := []struct {
		name    string
		logger  *slog.Logger
		repo    repository.TestcaseLocalStorageRepository
		wantErr bool
	}{
		{
			name:    "all not nil",
			logger:  logger,
			repo:    mockRepo,
			wantErr: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			repo:    mockRepo,
			wantErr: true,
		},
		{
			name:    "nil repo",
			logger:  logger,
			repo:    nil,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewTestcaseLocalStorageService(test.logger, test.repo)
			if (err != nil) != test.wantErr {
				t.Errorf("NewTestcaseStorageService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && service == nil {
				t.Errorf("NewTestcaseStorageService() returned nil service")
			}
		})
	}
}

func TestSave(t *testing.T) {
	logger := slog.Default()

	testCase := &entity.TestCase{
		TestID:      "test-1",
		Description: "Test description",
		TestCode: entity.TestCode{
			Code:     "console.log('test');",
			Language: "ts",
		},
		Status: entity.TestStatusPassed,
	}

	tests := []struct {
		name      string
		testcase  *entity.TestCase
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
	}{
		{
			name:      "successful save",
			testcase:  testCase,
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Save(testCase, "user-1", "session-1").Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "repository error",
			testcase:  testCase,
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Save(testCase, "user-1", "session-1").Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			err = service.Save(test.testcase, test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestGetTestPath(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		testId    string
		lang      string
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      string
	}{
		{
			name:      "successful get path",
			testId:    "test-1",
			lang:      "ts",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPath("test-1", "ts", "user-1", "session-1").Return("user-1/session-1/test-1.ts", nil)
			},
			wantErr: false,
			want:    "user-1/session-1/test-1.ts",
		},
		{
			name:      "repository error",
			testId:    "test-1",
			lang:      "ts",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPath("test-1", "ts", "user-1", "session-1").Return("", errors.New("repository error"))
			},
			wantErr: true,
			want:    "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			got, err := service.GetTestPath(test.testId, test.lang, test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("GetTestPath() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("GetTestPath() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetTestPathsBySession(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      []string
	}{
		{
			name:      "successful get paths with multiple files",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPathsBySession("user-1", "session-1").Return([]string{
					"user-1/session-1/test-1.ts",
					"user-1/session-1/test-2.ts",
				}, nil)
			},
			wantErr: false,
			want: []string{
				"user-1/session-1/test-1.ts",
				"user-1/session-1/test-2.ts",
			},
		},
		{
			name:      "successful get paths with empty result",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPathsBySession("user-1", "session-1").Return([]string{}, nil)
			},
			wantErr: false,
			want:    []string{},
		},
		{
			name:      "repository error",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPathsBySession("user-1", "session-1").Return(nil, errors.New("repository error"))
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			got, err := service.GetTestPathsBySession(test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("GetTestPathsBySession() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Errorf("GetTestPathsBySession() got %d paths, want %d", len(got), len(test.want))
			}
		})
	}
}

func TestGetTestPathsByUser(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		userId    string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      map[string][]string
	}{
		{
			name:   "successful get paths with multiple sessions",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				result := map[string][]string{
					"session-1": {"user-1/session-1/test-1.ts"},
					"session-2": {"user-1/session-2/test-2.ts"},
				}
				m.EXPECT().GetTestPathsByUser("user-1").Return(result, nil)
			},
			wantErr: false,
			want: map[string][]string{
				"session-1": {"user-1/session-1/test-1.ts"},
				"session-2": {"user-1/session-2/test-2.ts"},
			},
		},
		{
			name:   "successful get paths with empty result",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPathsByUser("user-1").Return(map[string][]string{}, nil)
			},
			wantErr: false,
			want:    map[string][]string{},
		},
		{
			name:   "repository error",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().GetTestPathsByUser("user-1").Return(nil, errors.New("repository error"))
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			got, err := service.GetTestPathsByUser(test.userId)
			if (err != nil) != test.wantErr {
				t.Errorf("GetTestPathsByUser() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Errorf("GetTestPathsByUser() got %d sessions, want %d", len(got), len(test.want))
			}
		})
	}
}

func TestDelete(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		testId    string
		lang      string
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
	}{
		{
			name:      "successful delete",
			testId:    "test-1",
			lang:      "ts",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Delete("test-1", "ts", "user-1", "session-1").Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "repository error",
			testId:    "test-1",
			lang:      "ts",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Delete("test-1", "ts", "user-1", "session-1").Return(errors.New("repository error"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			err = service.Delete(test.testId, test.lang, test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestCleanupOldTests(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
	}{
		{
			name: "successful cleanup",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().DeleteOlderThan(mock.Anything).Return(5, nil)
			},
			wantErr: false,
		},
		{
			name: "successful cleanup with no deleted items",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().DeleteOlderThan(mock.Anything).Return(0, nil)
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().DeleteOlderThan(mock.Anything).Return(0, errors.New("repository error"))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockTestcaseLocalStorageRepository(t)
			test.setupMock(mockRepo)

			service, err := NewTestcaseLocalStorageService(logger, mockRepo)
			if err != nil {
				t.Fatalf("NewTestcaseLocalStorageService() failed: %v", err)
			}

			err = service.CleanupOldTests()
			if (err != nil) != test.wantErr {
				t.Errorf("CleanupOldTests() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
