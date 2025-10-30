package repository

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

const (
	userID    = "user1"
	sessionID = "0001"
)

func TestNewTestcaseLocalStorageRepository(t *testing.T) {
	logger := slog.Default()
	filesystem := mocks.NewMockFileSystem(t)

	tests := []struct {
		name string
		log  *slog.Logger
		fs   FileSystem
		err  bool
	}{
		{
			name: "valid",
			log:  logger,
			fs:   filesystem,
			err:  false,
		},
		{
			name: "params nil",
			log:  nil,
			fs:   nil,
			err:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewTestcaseLocalStorageRepository(tt.log, tt.fs)
			if tt.err {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestSave(t *testing.T) {
	logger := slog.Default()

	baseTestcase := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Language: "en",
			Code:     "console.log('hello');",
		},
	}

	tests := []struct {
		name      string
		testcase  *entity.TestCase
		userID    string
		sessionID string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
	}{
		{
			name:      "success",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("MkdirAll", userID+"/"+sessionID).Return(nil).Once()
				m.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "mkdir fails",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("MkdirAll", userID+"/"+sessionID).Return(assert.AnError).Once()
				// WriteFile should not be called
			},
			wantErr: true,
		},
		{
			name:      "write fails",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("MkdirAll", userID+"/"+sessionID).Return(nil).Once()
				m.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name: "invalid test id (validation error)",
			testcase: &entity.TestCase{
				TestID: "invalid/test:id",
				TestCode: entity.TestCode{
					Language: "en",
					Code:     "console.log('bad');",
				},
			},
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				// validation fails before filesystem is called, so no expectations
			},
			wantErr: true,
		},
		{
			name:      "invalid path elements empty user/session",
			testcase:  baseTestcase,
			userID:    "",
			sessionID: "",
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			if tc.setupMock != nil {
				tc.setupMock(mockFS)
			}

			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			err = repo.Save(tc.testcase, tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
				// If validation prevented filesystem calls:
				if tc.name == "invalid test id (validation error)" || tc.userID == "" || tc.sessionID == "" {
					mockFS.AssertNotCalled(t, "MkdirAll")
					mockFS.AssertNotCalled(t, "WriteFile")
				}
			} else {
				assert.NoError(t, err)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestRead(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		testID    string
		lang      string
		userID    string
		sessionID string
		fileData  string
		setupMock func(m *mocks.MockFileSystem, filePath string)
		wantErr   bool
		want      *entity.TestCase
	}{
		{
			name:      "success",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "en",
			userID:    userID,
			sessionID: sessionID,
			fileData:  "console.log('test');",
			setupMock: func(m *mocks.MockFileSystem, filePath string) {
				m.On("ReadFile", filePath).Return([]byte("console.log('test');"), nil).Once()
			},
			wantErr: false,
			want: &entity.TestCase{
				TestID: "123e4567-e89b-12d3-a456-426614174000",
				TestCode: entity.TestCode{
					Language: "en",
					Code:     "console.log('test');",
				},
				Status: entity.TestStatusNotRun,
			},
		},
		{
			name:      "invalid filename",
			testID:    "invalid/test:id",
			lang:      "en",
			userID:    userID,
			sessionID: sessionID,
			setupMock: nil,
			wantErr:   true,
			want:      nil,
		},
		{
			name:      "invalid path elements empty user/session",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "en",
			userID:    "",
			sessionID: "",
			setupMock: nil,
			wantErr:   true,
			want:      nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			filePath := filepath.Join(tc.userID, tc.sessionID, tc.testID+"."+tc.lang)
			if tc.setupMock != nil {
				tc.setupMock(mockFS, filePath)
			}

			got, err := repo.Read(tc.testID, tc.lang, tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if tc.setupMock == nil {
					mockFS.AssertNotCalled(t, "ReadFile")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.want.TestID, got.TestID)
				assert.Equal(t, tc.want.TestCode.Language, got.TestCode.Language)
				assert.Equal(t, tc.want.TestCode.Code, got.TestCode.Code)
			}
			mockFS.AssertExpectations(t)
		})
	}
}

func TestReadAllbySession(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		userID    string
		sessionID string
		fileNames []string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
		wantCount int
	}{
		{
			name:      "success",
			userID:    userID,
			sessionID: sessionID,
			fileNames: []string{"123e4567-e89b-12d3-a456-426614174000.en", "819b4567-e89b-12d3-a456-426614174001.en"},
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return([]string{"123e4567-e89b-12d3-a456-426614174000.en", "819b4567-e89b-12d3-a456-426614174001.en"}, nil).Once()
				m.On("ReadFile", filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.en")).Return([]byte("console.log('123e4567-e89b-12d3-a456-426614174000');"), nil).Once()
				m.On("ReadFile", filepath.Join(userID, sessionID, "819b4567-e89b-12d3-a456-426614174001.en")).Return([]byte("console.log('819b4567-e89b-12d3-a456-426614174001');"), nil).Once()
			},
			wantErr:   false,
			wantCount: 2,
		},

		{
			name:      "read dir fails",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return(nil, assert.AnError).Once()
			},
			wantErr:   true,
			wantCount: 0,
		},

		{
			name:      "invalid testcase filename",
			userID:    userID,
			sessionID: sessionID,
			fileNames: []string{"invalid/test:id.en", "123e4567-e89b-12d3-a456-426614174000.en"},
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return([]string{"invalid/test:id.en", "123e4567-e89b-12d3-a456-426614174000.en"}, nil).Once()
				m.On("ReadFile", filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.en")).Return([]byte("console.log('123e4567-e89b-12d3-a456-426614174000');"), nil).Once()
			},
			wantErr:   false, // ReadAllBySession should skip invalid filenames but not error
			wantCount: 1,
		},

		{
			name:      "invalid path elements empty user/session",
			userID:    "",
			sessionID: "",
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			if tc.setupMock != nil {
				tc.setupMock(mockFS)
			}
			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			testcases, err := repo.ReadAllBySession(tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, testcases)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, testcases)
				assert.Equal(t, tc.wantCount, len(testcases))
			}
			mockFS.AssertExpectations(t)
		})
	}
}

func TestReadAllByUser(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		userID    string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
		wantCount int // total sessions in result map
	}{
		{
			name:   "success",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				// simulate two sessions for the user
				m.On("ReadDir", userID).Return([]string{"session1", "session2"}, nil).Once()

				// Each session directory should be read successfully
				m.On("ReadDir", filepath.Join(userID, "session1")).Return([]string{"123e4567-e89b-12d3-a456-426614174000.en"}, nil).Once()
				m.On("ReadFile", filepath.Join(userID, "session1", "123e4567-e89b-12d3-a456-426614174000.en")).Return([]byte("console.log('123e4567-e89b-12d3-a456-426614174000.en');"), nil).Once()

				m.On("ReadDir", filepath.Join(userID, "session2")).Return([]string{"819b4567-e89b-12d3-a456-426614174001.en"}, nil).Once()
				m.On("ReadFile", filepath.Join(userID, "session2", "819b4567-e89b-12d3-a456-426614174001.en")).Return([]byte("console.log('tc2');"), nil).Once()
			},
			wantErr:   false,
			wantCount: 2, // sessions
		},

		{
			name:   "invalid user id (empty)",
			userID: "",
			setupMock: func(m *mocks.MockFileSystem) {
				// nothing to mock, should fail at path validation
			},
			wantErr:   true,
			wantCount: 0,
		},

		{
			name:   "read user directory fails",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", userID).Return(nil, assert.AnError).Once()
			},
			wantErr:   true,
			wantCount: 0,
		},

		{
			name:   "read session fails",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				// user directory has one session
				m.On("ReadDir", userID).Return([]string{"session1"}, nil).Once()
				// simulate failure when reading that session’s testcases
				m.On("ReadDir", filepath.Join(userID, "session1")).Return(nil, assert.AnError).Once()
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			if tc.setupMock != nil {
				tc.setupMock(mockFS)
			}

			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			results, err := repo.ReadAllByUser(tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				assert.Equal(t, tc.wantCount, len(results))
			}

			mockFS.AssertExpectations(t)
		})
	}
}
