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
			Language: "javascript",
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

func TestRead(t *testing.T) {
	logger := slog.Default()

	testCase := &entity.TestCase{
		TestID:      "test-1",
		Description: "Test description",
		TestCode: entity.TestCode{
			Code:     "console.log('test');",
			Language: "javascript",
		},
		Status: entity.TestStatusPassed,
	}

	tests := []struct {
		name      string
		testId    string
		lang      string
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      *entity.TestCase
	}{
		{
			name:      "successful read",
			testId:    "test-1",
			lang:      "javascript",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Read("test-1", "javascript", "user-1", "session-1").Return(testCase, nil)
			},
			wantErr: false,
			want:    testCase,
		},
		{
			name:      "repository error",
			testId:    "test-1",
			lang:      "javascript",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Read("test-1", "javascript", "user-1", "session-1").Return(nil, errors.New("repository error"))
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

			got, err := service.Read(test.testId, test.lang, test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("Read() got = %v, want %v", got, test.want)
			}
		})
	}
}

func TestReadAllBySession(t *testing.T) {
	logger := slog.Default()

	testCase1 := &entity.TestCase{
		TestID:      "test-1",
		Description: "Test description 1",
		TestCode: entity.TestCode{
			Code:     "console.log('test1');",
			Language: "javascript",
		},
		Status: entity.TestStatusPassed,
	}

	testCase2 := &entity.TestCase{
		TestID:      "test-2",
		Description: "Test description 2",
		TestCode: entity.TestCode{
			Code:     "console.log('test2');",
			Language: "javascript",
		},
		Status: entity.TestStatusFailed,
	}

	tests := []struct {
		name      string
		userId    string
		sessionId string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      []*entity.TestCase
	}{
		{
			name:      "successful read with multiple testcases",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().ReadAllBySession("user-1", "session-1").Return([]*entity.TestCase{testCase1, testCase2}, nil)
			},
			wantErr: false,
			want:    []*entity.TestCase{testCase1, testCase2},
		},
		{
			name:      "successful read with empty result",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().ReadAllBySession("user-1", "session-1").Return([]*entity.TestCase{}, nil)
			},
			wantErr: false,
			want:    []*entity.TestCase{},
		},
		{
			name:      "repository error",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().ReadAllBySession("user-1", "session-1").Return(nil, errors.New("repository error"))
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

			got, err := service.ReadAllBySession(test.userId, test.sessionId)
			if (err != nil) != test.wantErr {
				t.Errorf("ReadAllBySession() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Errorf("ReadAllBySession() got %d testcases, want %d", len(got), len(test.want))
			}
		})
	}
}

func TestReadAllByUser(t *testing.T) {
	logger := slog.Default()

	testCase1 := &entity.TestCase{
		TestID:      "test-1",
		Description: "Test description 1",
		TestCode: entity.TestCode{
			Code:     "console.log('test1');",
			Language: "javascript",
		},
		Status: entity.TestStatusPassed,
	}

	testCase2 := &entity.TestCase{
		TestID:      "test-2",
		Description: "Test description 2",
		TestCode: entity.TestCode{
			Code:     "console.log('test2');",
			Language: "javascript",
		},
		Status: entity.TestStatusFailed,
	}

	tests := []struct {
		name      string
		userId    string
		setupMock func(*mocks.MockTestcaseLocalStorageRepository)
		wantErr   bool
		want      map[string][]*entity.TestCase
	}{
		{
			name:   "successful read with multiple sessions",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				result := map[string][]*entity.TestCase{
					"session-1": {testCase1},
					"session-2": {testCase2},
				}
				m.EXPECT().ReadAllByUser("user-1").Return(result, nil)
			},
			wantErr: false,
			want: map[string][]*entity.TestCase{
				"session-1": {testCase1},
				"session-2": {testCase2},
			},
		},
		{
			name:   "successful read with empty result",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().ReadAllByUser("user-1").Return(map[string][]*entity.TestCase{}, nil)
			},
			wantErr: false,
			want:    map[string][]*entity.TestCase{},
		},
		{
			name:   "repository error",
			userId: "user-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().ReadAllByUser("user-1").Return(nil, errors.New("repository error"))
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

			got, err := service.ReadAllByUser(test.userId)
			if (err != nil) != test.wantErr {
				t.Errorf("ReadAllByUser() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Errorf("ReadAllByUser() got %d sessions, want %d", len(got), len(test.want))
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
			lang:      "javascript",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Delete("test-1", "javascript", "user-1", "session-1").Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "repository error",
			testId:    "test-1",
			lang:      "javascript",
			userId:    "user-1",
			sessionId: "session-1",
			setupMock: func(m *mocks.MockTestcaseLocalStorageRepository) {
				m.EXPECT().Delete("test-1", "javascript", "user-1", "session-1").Return(errors.New("repository error"))
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
