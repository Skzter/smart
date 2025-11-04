package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
)

func ctx(t *testing.T) context.Context { return t.Context() }

func TestNewTaglistStorage(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		testName      string
		logger        *slog.Logger
		repo          func(*testing.T) repository.TaglistStorage
		expectedError bool
	}{
		{
			testName:      "Invalid Logger and Repo",
			logger:        nil,
			repo:          func(t *testing.T) repository.TaglistStorage { return nil },
			expectedError: true,
		},
		{
			testName:      "Invalid Repo Only",
			logger:        logger,
			repo:          func(t *testing.T) repository.TaglistStorage { return nil },
			expectedError: true,
		},
		{
			testName:      "Valid Parameters",
			logger:        logger,
			repo:          func(t *testing.T) repository.TaglistStorage { return mocks.NewMockTaglistStorage(t) },
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, err := NewTaglistStorage(test.logger, test.repo(t))

			if test.expectedError {
				assert.NotNil(t, err)
				assert.Nil(t, storage)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, storage)
			}
		})
	}
}

func TestStoreTaglist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		testName      string
		existsReturns *[]any
		createReturns *[]any
		updateReturns *[]any
		ctx           func(*testing.T) context.Context
		expectedError bool
	}{
		{
			testName:      "Nil Context",
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
		{
			testName:      "Repo TaglistExists Error",
			existsReturns: &[]any{false, errors.New("err")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:      "Create New Taglist Successfully",
			existsReturns: &[]any{false, nil},
			createReturns: &[]any{nil},
			ctx:           ctx,
			expectedError: false,
		},
		{
			testName:      "Update Existing Taglist Successfully",
			existsReturns: &[]any{true, nil},
			updateReturns: &[]any{nil},
			ctx:           ctx,
			expectedError: false,
		},
		{
			testName:      "Create Taglist Fails",
			existsReturns: &[]any{false, nil},
			createReturns: &[]any{errors.New("err")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:      "Update Taglist Fails",
			existsReturns: &[]any{true, nil},
			updateReturns: &[]any{errors.New("err")},
			ctx:           ctx,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			repo := mocks.NewMockTaglistStorage(t)
			if test.existsReturns != nil {
				repo.On("TaglistExists", mock.Anything).Return(*test.existsReturns...)
			}
			if test.createReturns != nil {
				repo.On("CreateTaglist", mock.Anything, mock.Anything).Return(*test.createReturns...)
			}
			if test.updateReturns != nil {
				repo.On("UpdateTaglist", mock.Anything, mock.Anything).Return(*test.updateReturns...)
			}

			storage, _ := NewTaglistStorage(logger, repo)

			err := storage.StoreTaglist(test.ctx(t), []string{"A", "B", "C"})

			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetTaglist(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		testName       string
		readReturns    *[]any
		ctx            func(t *testing.T) context.Context
		expectedError  bool
		expectedResult []string
	}{
		{
			testName:      "Nil Context",
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
		{
			testName:      "Repo ReadTaglist Error",
			readReturns:   &[]any{nil, errors.New("errors")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:       "Successful Taglist Read",
			readReturns:    &[]any{&entity.TagList{Tags: []string{"A", "B", "C"}}, nil},
			ctx:            ctx,
			expectedError:  false,
			expectedResult: []string{"A", "B", "C"},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			repo := mocks.NewMockTaglistStorage(t)

			if test.readReturns != nil {
				repo.On("ReadTaglist", mock.Anything).Return(*test.readReturns...)
			}

			storage, err := NewTaglistStorage(logger, repo)
			if err != nil {
				t.Fatalf("failed to init TaglistStorage: %v", err)
			}

			result, err := storage.GetTaglist(test.ctx(t))

			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedResult, result)
			}
		})
	}
}
