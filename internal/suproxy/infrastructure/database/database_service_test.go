package database

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
	mockRepo "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/mocks/repository"
)

func TestNewDatabaseService(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	repo := mockRepo.NewMockDatabaseRepository(t)
	tracer := otel.Tracer("test")

	svc, err := NewDatabaseService(logger, repo, tracer)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// nil logger
	svcNilLogger, err := NewDatabaseService(nil, repo, tracer)
	assert.Error(t, err)
	assert.Nil(t, svcNilLogger)
}

func TestDatabaseService_SaveDbEntry(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	repo := mockRepo.NewMockDatabaseRepository(t)
	tracer := otel.Tracer("test")

	svc, err := NewDatabaseService(logger, repo, tracer)
	assert.NoError(t, err)

	entry := entity.DatabaseEntry{
		Request: entity.Request{
			Header:      map[string]string{"Content-Type": "application/json"},
			Tags:        "Tags",
			Destination: "http://example.com",
			Body:        `{}`,
		},
		Response: entity.Response{Response: "OK"},
		Tags:     &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "TAG1"}, {Name: "TAG2"}}},
	}

	tests := []struct {
		name      string
		ctx       context.Context
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			ctx:  context.Background(),
			setupMock: func() {
				repo.On("CreateRequest", mock.Anything, entry).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "repo error",
			ctx:  context.Background(),
			setupMock: func() {
				repo.On("CreateRequest", mock.Anything, entry).Return(assert.AnError)
			},
			wantErr: true,
		},
		{
			name:      "nil context",
			ctx:       nil,
			setupMock: func() {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.ExpectedCalls = nil
			tt.setupMock()

			err := svc.SaveDbEntry(tt.ctx, entry)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestDatabaseService_GetAllKeys(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	repo := mockRepo.NewMockDatabaseRepository(t)
	tracer := otel.Tracer("test")

	svc, _ := NewDatabaseService(logger, repo, tracer)

	tests := []struct {
		name      string
		ctx       context.Context
		mockKeys  []string
		mockError error
		wantErr   bool
	}{
		{
			name:      "success",
			ctx:       context.Background(),
			mockKeys:  []string{"key1", "key2"},
			mockError: nil,
			wantErr:   false,
		},
		{
			name:      "repo error",
			ctx:       context.Background(),
			mockKeys:  nil,
			mockError: assert.AnError,
			wantErr:   true,
		},
		{
			name:      "nil context",
			ctx:       nil,
			mockKeys:  nil,
			mockError: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.ExpectedCalls = nil

			if tt.ctx != nil {
				repo.On("ListAllKeys", mock.Anything).Return(tt.mockKeys, tt.mockError)
			}

			keys, err := svc.GetAllKeys(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, keys)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockKeys, keys)
			}

			repo.AssertExpectations(t)
		})
	}
}
