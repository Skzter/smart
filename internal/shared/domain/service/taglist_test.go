package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
)

type mockCall struct {
	Name    string
	Params  int
	Returns []any
}

func call(name string, params int, returns ...any) *mockCall {
	return &mockCall{Name: name, Params: params, Returns: returns}
}

func taglistExists(b bool, err error) *mockCall { return call("TaglistExists", 1, b, err) }
func createTaglist(err error) *mockCall         { return call("CreateTaglist", 2, err) }
func updateTaglist(err error) *mockCall         { return call("UpdateTaglist", 2, err) }
func readTaglist(tle *entity.TagListEntity, err error) *mockCall {
	return call("ReadTaglist", 1, tle, err)
}

func setupMocks(calls ...*mockCall) func(*testing.T) repository.TaglistStorage {
	return func(t *testing.T) repository.TaglistStorage {
		repo := mocks.NewMockTaglistStorage(t)
		for _, call := range calls {
			args := make([]any, call.Params)
			for i := range args {
				args[i] = mock.Anything
			}
			repo.On(call.Name, args...).Return(call.Returns...)
		}
		return repo
	}
}

func ctx(t *testing.T) context.Context { return t.Context() }

func TestNewTaglistStorage(t *testing.T) {
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
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			repo:          func(t *testing.T) repository.TaglistStorage { return nil },
			expectedError: true,
		},
		{
			testName:      "Valid Parameters",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			repo:          setupMocks(),
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		testName      string
		repo          func(*testing.T) repository.TaglistStorage
		ctx           func(*testing.T) context.Context
		expectedError bool
	}{
		{
			testName:      "Nil Context",
			repo:          setupMocks(),
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
		{
			testName:      "Repo TaglistExists Error",
			repo:          setupMocks(taglistExists(false, errors.New("err"))),
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:      "Create New Taglist Successfully",
			repo:          setupMocks(taglistExists(false, nil), createTaglist(nil)),
			ctx:           ctx,
			expectedError: false,
		},
		{
			testName:      "Update Existing Taglist Successfully",
			repo:          setupMocks(taglistExists(true, nil), updateTaglist(nil)),
			ctx:           ctx,
			expectedError: false,
		},
		{
			testName:      "Create Taglist Fails",
			repo:          setupMocks(taglistExists(false, nil), createTaglist(errors.New("err"))),
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:      "Update Taglist Fails",
			repo:          setupMocks(taglistExists(true, nil), updateTaglist(errors.New("err"))),
			ctx:           ctx,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, _ := NewTaglistStorage(logger, test.repo(t))

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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		testName       string
		repo           func(t *testing.T) repository.TaglistStorage
		ctx            func(t *testing.T) context.Context
		expectedError  bool
		expectedResult []string
	}{
		{
			testName:      "Nil Context",
			repo:          setupMocks(),
			ctx:           func(t *testing.T) context.Context { return nil },
			expectedError: true,
		},
		{
			testName:      "Repo ReadTaglist Error",
			repo:          setupMocks(readTaglist(nil, errors.New("errors"))),
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:       "Successful Taglist Read",
			repo:           setupMocks(readTaglist(&entity.TagListEntity{Tags: []string{"A", "B", "C"}}, nil)),
			ctx:            ctx,
			expectedError:  false,
			expectedResult: []string{"A", "B", "C"},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, err := NewTaglistStorage(logger, test.repo(t))
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
