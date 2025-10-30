package repository

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

const (
	userID    = "user1"
	sessionID = "0001"
	testRoot  = "tempTests"
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
			testcase:  baseTestcase,
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
				assert.NotNil(t, paths)
				assert.Equal(t, tc.wantCount, len(paths))
				if tc.wantPaths != nil {
					assert.ElementsMatch(t, tc.wantPaths, paths)
				}
			}

			if tc.setupMock == nil {
				mockFS.AssertNotCalled(t, "ReadDir")
				mockFS.AssertNotCalled(t, "GetValidatedPath")
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

				session1Path := filepath.Join(userID, "session1")
				m.EXPECT().ReadDir(session1Path).Return([]string{
					"123e4567-e89b-12d3-a456-426614174000.ts",
				}, nil).Once()

				file1Path := filepath.Join(session1Path, "123e4567-e89b-12d3-a456-426614174000.ts")
				fullPath1 := filepath.Clean(filepath.Join(testRoot, file1Path))
				m.EXPECT().GetValidatedPath(file1Path).Return(fullPath1, nil).Once()

				session2Path := filepath.Join(userID, "session2")
				m.EXPECT().ReadDir(session2Path).Return([]string{
					"819b4567-e89b-12d3-a456-426614174001.ts",
				}, nil).Once()

				file2Path := filepath.Join(session2Path, "819b4567-e89b-12d3-a456-426614174001.ts")
				fullPath2 := filepath.Clean(filepath.Join(testRoot, file2Path))
				m.EXPECT().GetValidatedPath(file2Path).Return(fullPath2, nil).Once()
			},
			wantErr:      false,
			wantSessions: 2,
			wantPaths: map[string][]string{
				"session1": {
					filepath.Clean(filepath.Join(testRoot, "user1/session1/123e4567-e89b-12d3-a456-426614174000.ts")),
				},
				"session2": {
					filepath.Clean(filepath.Join(testRoot, "user1/session2/819b4567-e89b-12d3-a456-426614174001.ts")),
				},
			},
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
	// TODO
}

func TestDeleteOlderThan(t *testing.T) {
	// TODO
}
