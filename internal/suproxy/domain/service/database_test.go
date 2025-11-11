package service

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	mockRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/repository/mocks"
)

// TestNewDatabaseService tests the creation of a new DatabaseService instance.
func TestNewDatabaseService(t *testing.T) {
	validLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validRepo := mockRepo.NewMockDatabaseRepository(t)
	svc, err := NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	svcNilLogger, err := NewDatabaseService(nil, validRepo)
	assert.Error(t, err)
	assert.Nil(t, svcNilLogger)
}

// TestDatabaseServiceSaveDbEntry tests the SaveDbEntry method of the DatabaseService.
func TestDatabaseServiceSaveDbEntry(t *testing.T) {
	validLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	validRepo := mockRepo.NewMockDatabaseRepository(t)

	svc, err := NewDatabaseService(validLogger, validRepo)
	assert.NoError(t, err)

	entry := entity.DatabaseEntry{
		Request:  "Test request",
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
			name: "nil context",
			setupMock: func() {
				// No mock setup needed because the context is nil,
			},
			ctx:     nil,
			wantErr: true,
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
	validRepo := mockRepo.NewMockDatabaseRepository(t)
	svc, _ := NewDatabaseService(logger, validRepo)

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
			validRepo.ExpectedCalls = nil

			// Only set up mock if context is not nil, otherwise expect error from service
			if tt.ctx != nil {
				validRepo.On("ListAllKeys", mock.Anything).Return(tt.mockKeys, tt.mockError)
			}

			keys, err := svc.GetAllKeys(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockKeys, keys)
			}

			validRepo.AssertExpectations(t)
		})
	}
}
