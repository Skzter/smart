package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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
func (m *mockRepo) ListKeysFromFile(ctx context.Context) ([]string, error) {
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

	svc, err := NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	svcNilLogger, err := NewDatabaseService(nil, validRepo)
	assert.Error(t, err)
	assert.Nil(t, svcNilLogger)
}

// TestNewDatabaseServiceWithNilRepo tests the creation of a new DatabaseService with a nil repository.
func TestDatabaseServiceSaveDbEntry(t *testing.T) {
	validLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validRepo := new(mockRepo)

	svc, err := NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)

	entry := entity.DatabaseEntry{
		Request: entity.Request{
			Header:      []string{"Content-Type: application/json"},
			Prompt:      "prompt",
			Destination: "http://example.com",
			Request:     `{}`,
		},
		Response: entity.Response{Response: "OK"},
		Tags:     []string{"tag1", "tag2"},
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				validRepo.On("CreateRequest", mock.Anything, entry).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			setupMock: func() {
				validRepo.On("CreateRequest", mock.Anything, entry).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validRepo.ExpectedCalls = nil // reset expectations
			tt.setupMock()
			err := svc.SaveDbEntry(context.Background(), entry)
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
	svc, _ := NewDatabaseService(logger, mockRepo)

	tests := []struct {
		name      string
		mockKeys  []string
		mockError error
		wantErr   bool
	}{
		{
			name:      "success",
			mockKeys:  []string{"tag1-tag2-123456", "mock-test-98765"},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "repo error",
			mockKeys:  nil,
			mockError: assert.AnError,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.On("ListKeysFromFile", mock.Anything).Return(tt.mockKeys, tt.mockError)

			keys, err := svc.GetAllKeys(context.Background())

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
func TestDatabaseServiceGetKeysForTags(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	mockRepo := new(mockRepo)
	svc, _ := NewDatabaseService(logger, mockRepo)

	allKeys := []string{
		"mock-test-123456",
		"foo-bar-789012",
		"sample-data-555555",
	}

	tests := []struct {
		name      string
		tags      []string
		mockKeys  []string
		mockError error
		expected  []string
		wantErr   bool
	}{
		{
			name:      "matching key found",
			tags:      []string{"mock", "test"},
			mockKeys:  allKeys,
			mockError: nil,
			expected:  []string{"mock-test-123456"},
			wantErr:   false,
		},
		{
			name:      "no match",
			tags:      []string{"not", "found"},
			mockKeys:  allKeys,
			mockError: nil,
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "repo error",
			tags:      []string{"anything"},
			mockKeys:  nil,
			mockError: assert.AnError,
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.On("ListKeysFromFile", mock.Anything).Return(tt.mockKeys, tt.mockError)

			keys, err := svc.GetKeysForTags(context.Background(), tt.tags)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, keys)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
