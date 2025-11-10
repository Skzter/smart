package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/service"
)

// mockRepo is a mock implementation of the DatabaseRepository interface for testing purposes.
type mockRepo struct {
	mock.Mock
}

// NewMockRepo creates a new instance of mockRepo.
func (m *mockRepo) CreateRequest(ctx context.Context, entry entity.DatabaseEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

// ListKeysFromFile returns a list of keys from the mock repository.
func (m *mockRepo) ListAllKeys(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// DeleteRequest deletes a request from the mock repository.
func (m *mockRepo) DeleteRequest(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// ReadRequest retrieves a request from the mock repository by its key.
func (m *mockRepo) ReadRequest(ctx context.Context, key string) (*entity.DatabaseEntry, error) {
	args := m.Called(ctx, key)
	if entry, ok := args.Get(0).(*entity.DatabaseEntry); ok {
		return entry, args.Error(1)
	}
	return nil, args.Error(1)
}

// UpdateRequest updates a request in the mock repository.
func (m *mockRepo) UpdateRequest(ctx context.Context, key string, entry entity.DatabaseEntry) error {
	args := m.Called(ctx, key, entry)
	return args.Error(0)
}

// TestNewDatabaseService tests the creation of a new DatabaseService instance.
func TestNewDatabaseService(t *testing.T) {
	validLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validRepo := new(mockRepo)

	svc, err := service.NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	svcNilLogger, err := service.NewDatabaseService(nil, validRepo)
	assert.Error(t, err)
	assert.Nil(t, svcNilLogger)
}

// TestDatabaseServiceSaveDbEntry tests the SaveDbEntry method of the DatabaseService.
func TestDatabaseServiceSaveDbEntry(t *testing.T) {
	validLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validRepo := new(mockRepo)

	svc, err := service.NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)

	entry := entity.DatabaseEntry{
		Request: entity.Request{
			Tags:        "prompt",
			Destination: "http://example.com",
			Body:        `{}`,
		},
		Response: entity.Response{Response: "OK"},
		Tags:     []string{"tag1", "tag2"},
	}

	tests := []struct {
		name      string
		setupMock func()
		ctx       context.Context
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				validRepo.On("CreateRequest", mock.Anything, entry).Return(nil)
			},
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func() {
				validRepo.On("CreateRequest", mock.Anything, entry).Return(assert.AnError)
			},
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:      "nil context",
			setupMock: func() {},
			ctx:       nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validRepo.ExpectedCalls = nil // reset expectations
			tt.setupMock()
			err := svc.SaveDbEntry(tt.ctx, entry)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			validRepo.AssertExpectations(t)
		})
	}
}

// TestDatabaseServiceGetAllKeys tests the GetAllKeys method of the DatabaseService.
func TestDatabaseServiceGetAllKeys(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockRepo := new(mockRepo)
	svc, _ := service.NewDatabaseService(logger, mockRepo)

	tests := []struct {
		name      string
		mockKeys  []string
		mockError error
		wantErr   bool
		ctx       context.Context
	}{
		{
			name:      "success",
			mockKeys:  []string{"tag1-tag2-123456", "mock-test-98765"},
			mockError: nil,
			wantErr:   false,
			ctx:       context.Background(),
		},
		{
			name:      "repo error",
			mockKeys:  nil,
			mockError: assert.AnError,
			wantErr:   true,
			ctx:       context.Background(),
		},
		{
			name:      "nil context",
			mockKeys:  nil,
			mockError: nil,
			wantErr:   true,
			ctx:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil

			// Only set up mock if context is not nil, otherwise expect error from service
			if tt.ctx != nil {
				mockRepo.On("ListAllKeys", mock.Anything).Return(tt.mockKeys, tt.mockError)
			}

			keys, err := svc.GetAllKeys(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockKeys, keys)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
