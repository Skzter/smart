package repository

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

const (
	userID    = "user1"
	sessionID = "0001"
	testRoot  = "tempTests"
)

// testMockFileInfo implements os.FileInfo for tests
type testMockFileInfo struct {
	modTime time.Time
}

var _ os.FileInfo = testMockFileInfo{}

func (m testMockFileInfo) Name() string       { return "" }
func (m testMockFileInfo) Size() int64        { return 0 }
func (m testMockFileInfo) Mode() os.FileMode  { return 0 }
func (m testMockFileInfo) ModTime() time.Time { return m.modTime }
func (m testMockFileInfo) IsDir() bool        { return false }
func (m testMockFileInfo) Sys() interface{}   { return nil }

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

type saveTestCase struct {
	name      string
	testcase  *entity.TestCase
	userID    string
	sessionID string
	setupMock func(m *mocks.MockFileSystem)
	wantErr   bool
}

func getSaveTestCases() []saveTestCase {
	baseTestcase := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Language: "ts",
			Code:     "console.log('hello');",
		},
	}

	return []saveTestCase{
		{
			name:      "success",
			git stash			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				dirPath := filepath.Join(userID, sessionID)
				m.EXPECT().MkdirAll(dirPath).Return(nil).Once()

				expectedFilePath := filepath.Join(dirPath, baseTestcase.TestID+"."+baseTestcase.TestCode.Language)
				m.EXPECT().WriteFile(expectedFilePath, []byte(baseTestcase.TestCode.Code)).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "mkdir fails",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				dirPath := filepath.Join(userID, sessionID)
				m.EXPECT().MkdirAll(dirPath).Return(assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name:      "write fails",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				dirPath := filepath.Join(userID, sessionID)
				m.EXPECT().MkdirAll(dirPath).Return(nil).Once()
				expectedFilePath := filepath.Join(dirPath, baseTestcase.TestID+"."+baseTestcase.TestCode.Language)
				m.EXPECT().WriteFile(expectedFilePath, []byte(baseTestcase.TestCode.Code)).Return(assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name: "invalid test id (validation error)",
			testcase: &entity.TestCase{
				TestID: "invalid/test:id",
				TestCode: entity.TestCode{
					Language: "ts",
					Code:     "console.log('bad');",
				},
			},
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
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
}

func TestSave(t *testing.T) {
	logger := slog.Default()
	tests := getSaveTestCases()

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
		})
	}
}

type getTestPathTestCase struct {
	name      string
	testID    string
	lang      string
	userID    string
	sessionID string
	setupMock func(m *mocks.MockFileSystem, relativePath string)
	wantErr   bool
	wantPath  string
}

func getTestPathTestCases() []getTestPathTestCase {
	return []getTestPathTestCase{
		{
			name:      "success",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "ts",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem, relativePath string) {
				fullPath := filepath.Clean(filepath.Join(testRoot, relativePath))
				m.EXPECT().GetFileStats(relativePath).Return(nil, nil).Once()
				m.EXPECT().GetValidatedPath(relativePath).Return(fullPath, nil).Once()
			},
			wantErr:  false,
			wantPath: filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.ts")),
		},
		{
			name:      "file not found",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "ts",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem, relativePath string) {
				m.EXPECT().GetFileStats(relativePath).Return(nil, assert.AnError).Once()
			},
			wantErr:  true,
			wantPath: "",
		},
		{
			name:      "path validation fails",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "ts",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem, relativePath string) {
				m.EXPECT().GetFileStats(relativePath).Return(nil, nil).Once()
				m.EXPECT().GetValidatedPath(relativePath).Return("", assert.AnError).Once()
			},
			wantErr:  true,
			wantPath: "",
		},
		{
			name:      "invalid filename",
			testID:    "invalid/test:id",
			lang:      "ts",
			userID:    userID,
			sessionID: sessionID,
			setupMock: nil,
			wantErr:   true,
			wantPath:  "",
		},
		{
			name:      "invalid path elements empty user/session",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "ts",
			userID:    "",
			sessionID: "",
			setupMock: nil,
			wantErr:   true,
			wantPath:  "",
		},
	}
}

func TestGetTestPath(t *testing.T) {
	logger := slog.Default()
	tests := getTestPathTestCases()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			relativePath := filepath.Join(tc.userID, tc.sessionID, tc.testID+"."+tc.lang)
			if tc.setupMock != nil {
				tc.setupMock(mockFS, relativePath)
			}

			got, err := repo.GetTestPath(tc.testID, tc.lang, tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				if tc.setupMock == nil {
					mockFS.AssertNotCalled(t, "GetFileStats")
					mockFS.AssertNotCalled(t, "GetValidatedPath")
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
				assert.Equal(t, tc.wantPath, got)
			}
		})
	}
}

type getTestPathsBySessionTestCase struct {
	name      string
	userID    string
	sessionID string
	fileNames []string
	setupMock func(m *mocks.MockFileSystem)
	wantErr   bool
	wantCount int
	wantPaths []string
}

func getTestPathsBySessionTestCases() []getTestPathsBySessionTestCase {
	return []getTestPathsBySessionTestCase{
		{
			name:      "success",
			userID:    userID,
			sessionID: sessionID,
			fileNames: []string{"123e4567-e89b-12d3-a456-426614174000.ts", "819b4567-e89b-12d3-a456-426614174001.ts"},
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.ts",
					"819b4567-e89b-12d3-a456-426614174001.ts",
				}, nil).Once()

				path1 := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.ts")
				fullPath1 := filepath.Clean(filepath.Join(testRoot, path1))
				m.EXPECT().GetValidatedPath(path1).Return(fullPath1, nil).Once()

				path2 := filepath.Join(sessionPath, "819b4567-e89b-12d3-a456-426614174001.ts")
				fullPath2 := filepath.Clean(filepath.Join(testRoot, path2))
				m.EXPECT().GetValidatedPath(path2).Return(fullPath2, nil).Once()
			},
			wantErr:   false,
			wantCount: 2,
			wantPaths: []string{
				filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.ts")),
				filepath.Clean(filepath.Join(testRoot, "user1/0001/819b4567-e89b-12d3-a456-426614174001.ts")),
			},
		},
		{
			name:      "read dir fails",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return(nil, assert.AnError).Once()
			},
			wantErr:   true,
			wantCount: 0,
			wantPaths: nil,
		},
		{
			name:      "invalid testcase filename - should skip",
			userID:    userID,
			sessionID: sessionID,
			fileNames: []string{"invalid/test:id.ts", "123e4567-e89b-12d3-a456-426614174000.ts"},
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"invalid/test:id.ts",
					"123e4567-e89b-12d3-a456-426614174000.ts",
				}, nil).Once()

				validPath := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.ts")
				fullPath := filepath.Clean(filepath.Join(testRoot, validPath))
				m.EXPECT().GetValidatedPath(validPath).Return(fullPath, nil).Once()
			},
			wantErr:   false,
			wantCount: 1,
			wantPaths: []string{
				filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.ts")),
			},
		},
		{
			name:      "path validation fails",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.ts",
				}, nil).Once()

				validPath := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.ts")
				m.EXPECT().GetValidatedPath(validPath).Return("", assert.AnError).Once()
			},
			wantErr:   true,
			wantCount: 0,
			wantPaths: nil,
		},
		{
			name:      "invalid path elements empty user/session",
			userID:    "",
			sessionID: "",
			setupMock: nil,
			wantErr:   true,
			wantCount: 0,
			wantPaths: nil,
		},
	}
}

func TestGetTestPathsBySession(t *testing.T) {
	logger := slog.Default()
	tests := getTestPathsBySessionTestCases()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			if tc.setupMock != nil {
				tc.setupMock(mockFS)
			}
			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			paths, err := repo.GetTestPathsBySession(tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, paths)
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
			name:   "success with multiple sessions",
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
			wantErr:      true,
			wantSessions: 0,
			wantPaths:    nil,
		},
		{
			name:   "read user directory fails",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(userID).Return(nil, assert.AnError).Once()
			},
			wantErr:      true,
			wantSessions: 0,
			wantPaths:    nil,
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
			wantErr:      false,
			wantSessions: 1,
			wantPaths: map[string][]string{
				"session1": {},
			},
		},
	}
}

func TestGetTestPathsByUser(t *testing.T) {
	logger := slog.Default()
	tests := getTestPathsByUserTestCases()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := mocks.NewMockFileSystem(t)
			if tc.setupMock != nil {
				tc.setupMock(mockFS)
			}

			repo, err := NewTestcaseLocalStorageRepository(logger, mockFS)
			assert.NoError(t, err)
			assert.NotNil(t, repo)

			results, err := repo.GetTestPathsByUser(tc.userID)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, results)
				assert.Equal(t, tc.wantSessions, len(results))
				if tc.wantPaths != nil {
					for sessionId, expectedPaths := range tc.wantPaths {
						actualPaths, ok := results[sessionId]
						assert.True(t, ok, "session %s should exist in results", sessionId)
						assert.ElementsMatch(t, expectedPaths, actualPaths)
					}
				}
			}

			if tc.userID == "" {
				mockFS.AssertNotCalled(t, "ReadDir")
				mockFS.AssertNotCalled(t, "GetValidatedPath")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	// TODO
}

func TestDeleteOlderThan(t *testing.T) {
	// TODO
}

func TestDelete(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		testID    string
		lang      string
		userID    string
		sessionID string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
	}{
		{
			name:      "success",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "en",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				path := filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.en")
				m.On("Remove", path, false).Return(nil).Once()
			},
			wantErr: false,
		},

		{
			name:      "invalid path elements",
			testID:    "",
			lang:      "en",
			userID:    "",
			sessionID: "",
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
		},

		{
			name:      "invalid filename",
			testID:    "invalid:test",
			lang:      "en",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
		},

		{
			name:      "remove fails",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			lang:      "en",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				path := filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.en")
				m.On("Remove", path, false).Return(assert.AnError).Once()
			},
			wantErr: true,
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

			err = repo.Delete(tc.testID, tc.lang, tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestDeleteOlderThan(t *testing.T) {
	logger := slog.Default()

	now := time.Now()
	oldTime := now.Add(-48 * time.Hour)

	tests := []struct {
		name        string
		setupMock   func(m *mocks.MockFileSystem)
		wantErr     bool
		wantDeleted int
	}{
		{
			name: "success - deletes one old file",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", ".").Return([]string{userID}, nil).Once()
				m.On("ReadDir", userID).Return([]string{sessionID}, nil).Once()
				fname := "123e4567-e89b-12d3-a456-426614174000.en"
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return([]string{fname}, nil).Once()

				filePath := filepath.Join(userID, sessionID, fname)
				m.On("GetFileStats", filePath).Return(testMockFileInfo{modTime: oldTime}, nil).Once()

				m.On("Remove", filePath, false).Return(nil).Once()
			},
			wantErr:     false,
			wantDeleted: 1,
		},

		{
			name: "read user dir fails",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", ".").Return(nil, assert.AnError).Once()
			},
			wantErr:     true,
			wantDeleted: 0,
		},

		{
			name: "no files older than cutoff",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", ".").Return([]string{userID}, nil).Once()
				m.On("ReadDir", userID).Return([]string{sessionID}, nil).Once()
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return([]string{"newfile.en"}, nil).Once()

				filePath := filepath.Join(userID, sessionID, "newfile.en")
				m.On("GetFileStats", filePath).Return(testMockFileInfo{modTime: now}, nil).Once()
			},
			wantErr:     false,
			wantDeleted: 0,
		},

		{
			name: "delete fails but continues",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadDir", ".").Return([]string{userID}, nil).Once()
				m.On("ReadDir", userID).Return([]string{sessionID}, nil).Once()
				fname := "123e4567-e89b-12d3-a456-426614174000.en"
				m.On("ReadDir", filepath.Join(userID, sessionID)).Return([]string{fname}, nil).Once()

				filePath := filepath.Join(userID, sessionID, fname)
				m.On("GetFileStats", filePath).Return(testMockFileInfo{modTime: oldTime}, nil).Once()
				m.On("Remove", filePath, false).Return(assert.AnError).Once()
			},
			wantErr:     false,
			wantDeleted: 0,
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

			deleted, err := repo.DeleteOlderThan(24 * time.Hour)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Equal(t, 0, deleted)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantDeleted, deleted)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestReadTestcaseFromPath(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name      string
		path      string
		testID    string
		lang      string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
		wantCode  string
	}{
		{
			name:   "success",
			path:   "user1/session1/123e4567-e89b-12d3-a456-426614174000.en",
			testID: "123e4567-e89b-12d3-a456-426614174000",
			lang:   "en",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadFile", "user1/session1/123e4567-e89b-12d3-a456-426614174000.en").
					Return([]byte("console.log('hello');"), nil).
					Once()
			},
			wantErr:  false,
			wantCode: "console.log('hello');",
		},
		{
			name:   "read file fails",
			path:   "user1/session1/123e4567-e89b-12d3-a456-426614174000.en",
			testID: "123e4567-e89b-12d3-a456-426614174000",
			lang:   "en",
			setupMock: func(m *mocks.MockFileSystem) {
				m.On("ReadFile", "user1/session1/123e4567-e89b-12d3-a456-426614174000.en").
					Return(nil, assert.AnError).
					Once()
			},
			wantErr:  true,
			wantCode: "",
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

			//FIXME:
			// NewTestcaseLocalStorageRepository returns the interface type; we need the
			// concrete implementation to call the unexported helper readTestcaseFromPath.
			concreteRepo, ok := repo.(*testcaseLocalStorageRepository)
			if !ok {
				t.Fatalf("unexpected repo type: %T", repo)
			}

			result, err := concreteRepo.readTestcaseFromPath(tc.path, tc.testID, tc.lang)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.testID, result.TestID)
				assert.Equal(t, tc.lang, result.TestCode.Language)
				assert.Equal(t, tc.wantCode, result.TestCode.Code)
				assert.Equal(t, entity.TestStatusNotRun, result.Status)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestValidateTestcase(t *testing.T) {
	valid := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Code:     "print('ok')",
			Language: "py",
		},
	}

	tests := []struct {
		name      string
		testcase  *entity.TestCase
		wantError bool
	}{
		{"valid testcase",
			valid,
			false},

		{"nil testcase",
			nil,
			true},

		{"empty TestID",
			&entity.TestCase{TestCode: valid.TestCode},
			true},

		{"empty Code",
			&entity.TestCase{TestID: valid.TestID, TestCode: entity.TestCode{Language: "py"}},
			true},

		{"empty Language",
			&entity.TestCase{TestID: valid.TestID, TestCode: entity.TestCode{Code: "console.log('ok')"}},
			true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTestcase(tc.testcase)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePathNameElements(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		wantErr bool
	}{
		{
			name:    "all valid elements",
			values:  []string{userID, sessionID, "123e4567-e89b-12d3-a456-426614174000"},
			wantErr: false,
		},
		{
			name:    "single valid element",
			values:  []string{userID},
			wantErr: false,
		},
		{
			name:    "contains empty element",
			values:  []string{userID, "", "123e4567-e89b-12d3-a456-426614174000"},
			wantErr: true,
		},
		{
			name:    "all empty",
			values:  []string{"", "", ""},
			wantErr: true,
		},
		{
			name:    "no elements provided",
			values:  []string{},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePathNameElements(tc.values...)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	validID := "123e4567-e89b-12d3-a456-426614174000"

	tests := []struct {
		name         string
		filename     string
		wantErr      bool
		wantTestID   string
		wantLanguage string
	}{
		{"valid filename", validID + ".go", false, validID, "go"},
		{"empty filename", "", true, "", ""},
		{"missing extension", validID, true, "", ""},
		{"empty extension", validID + ".", true, "", ""},
		{"invalid UUID", "not-a-uuid.go", true, "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testID, lang, err := validateFilename(tc.filename)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTestID, testID)
				assert.Equal(t, tc.wantLanguage, lang)
			}
		})
	}
}
}
