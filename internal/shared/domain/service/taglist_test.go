package service_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

type mockTaglistRepo struct {
	existsReturn bool
	existsError  error
	createCalled bool
	createError  error
	updateCalled bool
	updateError  error
	readReturn   entity.TagListEntity
	readError    error
}

func (m *mockTaglistRepo) TaglistExists(ctx context.Context) (bool, error) {
	return m.existsReturn, m.existsError
}
func (m *mockTaglistRepo) CreateTaglist(ctx context.Context, taglist entity.TagListEntity) error {
	m.createCalled = true
	return m.createError
}
func (m *mockTaglistRepo) UpdateTaglist(ctx context.Context, taglist entity.TagListEntity) error {
	m.updateCalled = true
	return m.updateError
}
func (m *mockTaglistRepo) ReadTaglist(ctx context.Context) (*entity.TagListEntity, error) {
	return &m.readReturn, m.readError
}

func NewMockTaglistStorage() repository.TaglistStorage {
	return &mockTaglistRepo{}
}

func TestNewTaglistStorage(t *testing.T) {
	tests := []struct {
		testName      string
		logger        *slog.Logger
		repo          repository.TaglistStorage
		expectedError bool
	}{
		{
			testName:      "Invalid Logger and Repo",
			logger:        nil,
			repo:          nil,
			expectedError: true,
		},
		{
			testName:      "Invalid Repo Only",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			repo:          nil,
			expectedError: true,
		},
		{
			testName:      "Valid Parameters",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			repo:          NewMockTaglistStorage(),
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, err := service.NewTaglistStorage(test.logger, test.repo)

			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if storage != nil {
					t.Error("expected storage to be nil, got instance")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if storage == nil {
					t.Error("expected valid TaglistStorage, got nil")
				}
			}
		})
	}
}

func TestStoreTaglist(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		testName       string
		repo           *mockTaglistRepo
		ctx            context.Context
		expectedError  bool
		expectedCreate bool
		expectedUpdate bool
	}{
		{
			testName:      "Nil Context",
			repo:          &mockTaglistRepo{},
			ctx:           nil,
			expectedError: true,
		},
		{
			testName:      "Repo TaglistExists Error",
			repo:          &mockTaglistRepo{existsError: fmt.Errorf("s3 connection failed")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:       "Create New Taglist Successfully",
			repo:           &mockTaglistRepo{existsReturn: false},
			ctx:            ctx,
			expectedError:  false,
			expectedCreate: true,
		},
		{
			testName:       "Update Existing Taglist Successfully",
			repo:           &mockTaglistRepo{existsReturn: true},
			ctx:            ctx,
			expectedError:  false,
			expectedUpdate: true,
		},
		{
			testName:      "Create Taglist Fails",
			repo:          &mockTaglistRepo{existsReturn: false, createError: fmt.Errorf("create failed")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:      "Update Taglist Fails",
			repo:          &mockTaglistRepo{existsReturn: true, updateError: fmt.Errorf("update failed")},
			ctx:           ctx,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, err := service.NewTaglistStorage(logger, test.repo)
			if err != nil {
				t.Fatalf("failed to init TaglistStorage: %v", err)
			}

			err = storage.StoreTaglist(test.ctx, []string{"A", "B", "C"})

			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if test.expectedCreate && !test.repo.createCalled {
				t.Error("expected CreateTaglist to be called")
			}
			if test.expectedUpdate && !test.repo.updateCalled {
				t.Error("expected UpdateTaglist to be called")
			}
		})
	}
}

func TestGetTaglist(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		testName       string
		repo           *mockTaglistRepo
		ctx            context.Context
		expectedError  bool
		expectedResult []string
	}{
		{
			testName:      "Nil Context",
			repo:          &mockTaglistRepo{},
			ctx:           nil,
			expectedError: true,
		},
		{
			testName:      "Repo ReadTaglist Error",
			repo:          &mockTaglistRepo{readError: fmt.Errorf("read failed")},
			ctx:           ctx,
			expectedError: true,
		},
		{
			testName:       "Successful Taglist Read",
			repo:           &mockTaglistRepo{readReturn: entity.TagListEntity{Tags: []string{"A", "B", "C"}}},
			ctx:            ctx,
			expectedError:  false,
			expectedResult: []string{"A", "B", "C"},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			storage, err := service.NewTaglistStorage(logger, test.repo)
			if err != nil {
				t.Fatalf("failed to init TaglistStorage: %v", err)
			}

			result, err := storage.GetTaglist(test.ctx)

			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if len(result) != len(test.expectedResult) {
					t.Errorf("expected %v, got %v", test.expectedResult, result)
				}
			}
		})
	}
}
