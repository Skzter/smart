package repository

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
)

const (
	userID    = "user1"
	sessionID = "session1"
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

func TestSaveSuccessful(t *testing.T) {
	logger := slog.Default()
	filesystem := mocks.NewMockFileSystem(t)

	repo, err := NewTestcaseLocalStorageRepository(logger, filesystem)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	testcase := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Language: "en",
			Code:     "console.log('hello');",
		},
	}

	// Mock directory creation
	filesystem.On("MkdirAll", userID+"/"+sessionID).Return(nil).Once()

	// Mock file writing
	filesystem.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err = repo.Save(testcase, userID, sessionID)
	assert.NoError(t, err)

	filesystem.AssertExpectations(t)
}

func TestSaveFalseMkdir(t *testing.T) {
	logger := slog.Default()
	filesystem := mocks.NewMockFileSystem(t)

	repo, err := NewTestcaseLocalStorageRepository(logger, filesystem)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	testcase := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Language: "en",
			Code:     "console.log('hello');",
		},
	}

	// Mock directory creation failure
	filesystem.On("MkdirAll", userID+"/"+sessionID).Return(assert.AnError).Once()

	err = repo.Save(testcase, userID, sessionID)
	assert.Error(t, err)

	filesystem.AssertExpectations(t)
}

func TestSaveInvalidTestID(t *testing.T) {
	logger := slog.Default()
	filesystem := mocks.NewMockFileSystem(t)

	repo, err := NewTestcaseLocalStorageRepository(logger, filesystem)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	testcase := &entity.TestCase{
		TestID: "invalid/test:id",
		TestCode: entity.TestCode{
			Language: "en",
			Code:     "console.log('hello');",
		},
	}

	err = repo.Save(testcase, userID, sessionID)
	assert.Error(t, err)
}

func TestSaveInvalidPathElements(t *testing.T) {
	logger := slog.Default()
	filesystem := mocks.NewMockFileSystem(t)

	repo, err := NewTestcaseLocalStorageRepository(logger, filesystem)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	testcase := &entity.TestCase{
		TestID: "123e4567-e89b-12d3-a456-426614174000",
		TestCode: entity.TestCode{
			Language: "en",
			Code:     "console.log('hello');",
		},
	}

	// Case 1: empty userId
	err = repo.Save(testcase, "", sessionID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validate path elements")

	// Case 2: empty sessionId
	err = repo.Save(testcase, userID, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validate path elements")

	// Check if no filesystem methods were called
	filesystem.AssertNotCalled(t, "MkdirAll")
	filesystem.AssertNotCalled(t, "WriteFile")
}
