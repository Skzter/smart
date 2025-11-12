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
			Code: "console.log('hello');",
		},
	}

	return []saveTestCase{
		{
			name:      "success",
			testcase:  baseTestcase,
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				dirPath := filepath.Join(userID, sessionID)
				m.EXPECT().MkdirAll(dirPath).Return(nil).Once()

				expectedFilePath := filepath.Join(dirPath, baseTestcase.TestID+"."+"spec.ts")
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
				expectedFilePath := filepath.Join(dirPath, baseTestcase.TestID+"."+"spec.ts")
				m.EXPECT().WriteFile(expectedFilePath, []byte(baseTestcase.TestCode.Code)).Return(assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name: "invalid test id (validation error)",
			testcase: &entity.TestCase{
				TestID: "invalid/test:id",
				TestCode: entity.TestCode{
					Code: "console.log('bad');",
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
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem, relativePath string) {
				fullPath := filepath.Clean(filepath.Join(testRoot, relativePath))
				m.EXPECT().GetFileStats(relativePath).Return(nil, nil).Once()
				m.EXPECT().GetValidatedPath(relativePath).Return(fullPath, nil).Once()
			},
			wantErr:  false,
			wantPath: filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.spec.ts")),
		},
		{
			name:      "file not found",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
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
			userID:    userID,
			sessionID: sessionID,
			setupMock: nil,
			wantErr:   true,
			wantPath:  "",
		},
		{
			name:      "invalid path elements empty user/session",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
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

			relativePath := filepath.Join(tc.userID, tc.sessionID, tc.testID+".spec.ts")
			if tc.setupMock != nil {
				tc.setupMock(mockFS, relativePath)
			}

			got, err := repo.GetTestPath(tc.testID, tc.userID, tc.sessionID)
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
			fileNames: []string{"123e4567-e89b-12d3-a456-426614174000.spec.ts", "819b4567-e89b-12d3-a456-426614174001.spec.ts"},
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.spec.ts",
					"819b4567-e89b-12d3-a456-426614174001.spec.ts",
				}, nil).Once()

				path1 := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
				fullPath1 := filepath.Clean(filepath.Join(testRoot, path1))
				m.EXPECT().GetValidatedPath(path1).Return(fullPath1, nil).Once()

				path2 := filepath.Join(sessionPath, "819b4567-e89b-12d3-a456-426614174001.spec.ts")
				fullPath2 := filepath.Clean(filepath.Join(testRoot, path2))
				m.EXPECT().GetValidatedPath(path2).Return(fullPath2, nil).Once()
			},
			wantErr:   false,
			wantCount: 2,
			wantPaths: []string{
				filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.spec.ts")),
				filepath.Clean(filepath.Join(testRoot, "user1/0001/819b4567-e89b-12d3-a456-426614174001.spec.ts")),
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
			fileNames: []string{"invalid/test:id.spec.ts", "123e4567-e89b-12d3-a456-426614174000.spec.ts"},
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"invalid/test:id.spec.ts",
					"123e4567-e89b-12d3-a456-426614174000.spec.ts",
				}, nil).Once()

				validPath := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
				fullPath := filepath.Clean(filepath.Join(testRoot, validPath))
				m.EXPECT().GetValidatedPath(validPath).Return(fullPath, nil).Once()
			},
			wantErr:   false,
			wantCount: 1,
			wantPaths: []string{
				filepath.Clean(filepath.Join(testRoot, "user1/0001/123e4567-e89b-12d3-a456-426614174000.spec.ts")),
			},
		},
		{
			name:      "path validation fails",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				sessionPath := filepath.Join(userID, sessionID)
				m.EXPECT().ReadDir(sessionPath).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.spec.ts",
				}, nil).Once()

				validPath := filepath.Join(sessionPath, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
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
				assert.NotNil(t, paths)
				assert.Equal(t, tc.wantCount, len(paths))
			}
		})
	}
}

type getTestPathsByUserTestCase struct {
	name         string
	userID       string
	setupMock    func(m *mocks.MockFileSystem)
	wantErr      bool
	wantSessions int
	wantPaths    map[string][]string
}

func getTestPathsByUserTestCases() []getTestPathsByUserTestCase {
	return []getTestPathsByUserTestCase{
		{
			name:   "success with multiple sessions",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(userID).Return([]string{"session1", "session2"}, nil).Once()

				// session1 files
				session1Path := filepath.Join(userID, "session1")
				m.EXPECT().ReadDir(session1Path).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.spec.ts",
					"223e4567-e89b-12d3-a456-426614174001.spec.ts",
				}, nil).Once()

				path1 := filepath.Join(session1Path, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
				fullPath1 := filepath.Clean(filepath.Join(testRoot, path1))
				m.EXPECT().GetValidatedPath(path1).Return(fullPath1, nil).Once()

				path2 := filepath.Join(session1Path, "223e4567-e89b-12d3-a456-426614174001.spec.ts")
				fullPath2 := filepath.Clean(filepath.Join(testRoot, path2))
				m.EXPECT().GetValidatedPath(path2).Return(fullPath2, nil).Once()

				// session2 files
				session2Path := filepath.Join(userID, "session2")
				m.EXPECT().ReadDir(session2Path).Return([]string{
					"323e4567-e89b-12d3-a456-426614174002.spec.ts",
				}, nil).Once()

				path3 := filepath.Join(session2Path, "323e4567-e89b-12d3-a456-426614174002.spec.ts")
				fullPath3 := filepath.Clean(filepath.Join(testRoot, path3))
				m.EXPECT().GetValidatedPath(path3).Return(fullPath3, nil).Once()
			},
			wantErr:      false,
			wantSessions: 2,
			wantPaths: map[string][]string{
				"session1": {
					filepath.Clean(filepath.Join(testRoot, "user1/session1/123e4567-e89b-12d3-a456-426614174000.spec.ts")),
					filepath.Clean(filepath.Join(testRoot, "user1/session1/223e4567-e89b-12d3-a456-426614174001.spec.ts")),
				},
				"session2": {
					filepath.Clean(filepath.Join(testRoot, "user1/session2/323e4567-e89b-12d3-a456-426614174002.spec.ts")),
				},
			},
		},
		{
			name:         "empty user id",
			userID:       "",
			setupMock:    func(m *mocks.MockFileSystem) {},
			wantErr:      true,
			wantSessions: 0,
			wantPaths:    nil,
		},
		{
			name:   "read dir fails",
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
				m.EXPECT().ReadDir(userID).Return([]string{"session1"}, nil).Once()

				session1Path := filepath.Join(userID, "session1")
				m.EXPECT().ReadDir(session1Path).Return(nil, assert.AnError).Once()
			},
			wantErr:      true,
			wantSessions: 0,
			wantPaths:    nil,
		},
		{
			name:   "empty session",
			userID: userID,
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(userID).Return([]string{"session1"}, nil).Once()

				session1Path := filepath.Join(userID, "session1")
				m.EXPECT().ReadDir(session1Path).Return([]string{}, nil).Once()
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
	logger := slog.Default()

	tests := []struct {
		name      string
		testID    string
		userID    string
		sessionID string
		setupMock func(m *mocks.MockFileSystem)
		wantErr   bool
	}{
		{
			name:      "success",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				path := filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
				m.EXPECT().Remove(path, false).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:      "invalid path elements",
			testID:    "",
			userID:    "",
			sessionID: "",
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
		},
		{
			name:      "invalid filename",
			testID:    "invalid:test",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {},
			wantErr:   true,
		},
		{
			name:      "remove fails",
			testID:    "123e4567-e89b-12d3-a456-426614174000",
			userID:    userID,
			sessionID: sessionID,
			setupMock: func(m *mocks.MockFileSystem) {
				path := filepath.Join(userID, sessionID, "123e4567-e89b-12d3-a456-426614174000.spec.ts")
				m.EXPECT().Remove(path, false).Return(assert.AnError).Once()
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

			err = repo.Delete(tc.testID, tc.userID, tc.sessionID)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
				m.EXPECT().ReadDir(".").Return([]string{userID}, nil).Once()
				m.EXPECT().ReadDir(userID).Return([]string{sessionID}, nil).Once()
				fname := "123e4567-e89b-12d3-a456-426614174000.spec.ts"
				m.EXPECT().ReadDir(filepath.Join(userID, sessionID)).Return([]string{fname}, nil).Once()

				filePath := filepath.Join(userID, sessionID, fname)
				m.EXPECT().GetFileStats(filePath).Return(testMockFileInfo{modTime: oldTime}, nil).Once()
				m.EXPECT().Remove(filePath, false).Return(nil).Once()
			},
			wantErr:     false,
			wantDeleted: 1,
		},
		{
			name: "read user dir fails",
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(".").Return(nil, assert.AnError).Once()
			},
			wantErr:     true,
			wantDeleted: 0,
		},
		{
			name: "no files older than cutoff",
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(".").Return([]string{userID}, nil).Once()
				m.EXPECT().ReadDir(userID).Return([]string{sessionID}, nil).Once()
				m.EXPECT().ReadDir(filepath.Join(userID, sessionID)).Return([]string{"newfile"}, nil).Once()

				filePath := filepath.Join(userID, sessionID, "newfile")
				m.EXPECT().GetFileStats(filePath).Return(testMockFileInfo{modTime: now}, nil).Once()
			},
			wantErr:     false,
			wantDeleted: 0,
		},
		{
			name: "delete fails but continues",
			setupMock: func(m *mocks.MockFileSystem) {
				m.EXPECT().ReadDir(".").Return([]string{userID}, nil).Once()
				m.EXPECT().ReadDir(userID).Return([]string{sessionID}, nil).Once()
				fname := "123e4567-e89b-12d3-a456-426614174000.spec.ts"
				m.EXPECT().ReadDir(filepath.Join(userID, sessionID)).Return([]string{fname}, nil).Once()

				filePath := filepath.Join(userID, sessionID, fname)
				m.EXPECT().GetFileStats(filePath).Return(testMockFileInfo{modTime: oldTime}, nil).Once()
				m.EXPECT().Remove(filePath, false).Return(assert.AnError).Once()
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
		})
	}
}

func TestValidateTestcase(t *testing.T) {
	valid := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Code: "print('ok')",
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
			&entity.TestCase{TestID: valid.TestID, TestCode: entity.TestCode{Code: ""}}, true},
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
		name     string
		filename string
		wantErr  bool
	}{
		{"valid filename", validID + ".spec.ts", false},
		{"empty filename", "", true},
		{"missing extension", validID, true},
		{"empty extension", validID + ".", true},
		{"invalid UUID", "not-a-uuid.spec.ts", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFilename(tc.filename)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
