package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	servmocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

// nolint: dupl
func TestNewGroupStorage(t *testing.T) {
	logger := slog.Default()
	mockRepo := mocks.NewMockGroupStorage(t)
	mockValidator := servmocks.NewMockValidator(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		logger    *slog.Logger
		repo      repository.GroupStorage
		validator Validator
		tracer    trace.Tracer
		wantErr   bool
	}{
		{
			name:      "all not nil",
			logger:    logger,
			repo:      mockRepo,
			validator: mockValidator,
			tracer:    tracer,
			wantErr:   false,
		},
		{
			name:      "nil logger",
			logger:    nil,
			repo:      mockRepo,
			validator: mockValidator,
			tracer:    tracer,
			wantErr:   true,
		},
		{
			name:      "nil repo",
			logger:    logger,
			repo:      nil,
			validator: mockValidator,
			tracer:    tracer,
			wantErr:   true,
		},
		{
			name:      "nil validator",
			logger:    logger,
			repo:      mockRepo,
			validator: nil,
			tracer:    tracer,
			wantErr:   true,
		},
		{
			name:      "nil tracer",
			logger:    logger,
			repo:      mockRepo,
			validator: mockValidator,
			tracer:    nil,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svc, err := NewGroupStorage(test.logger, test.repo, test.validator, test.tracer)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, svc, "service should be nil on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, svc, "service should not be nil")
			}
		})
	}
}

// nolint: dupl
func TestGroupStorageNew(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name            string
		ctx             context.Context
		group           *entity.Group
		validReturns    []any
		createReturns   []any
		wantErr         bool
		expectedValCall int
		expectedRepCall int
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Test Group"},
			validReturns:    []any{nil},
			createReturns:   []any{nil},
			wantErr:         false,
			expectedValCall: 1,
			expectedRepCall: 1,
		},
		{
			name:            "nil context",
			ctx:             nil,
			group:           &entity.Group{Id: "group1", Name: "Test Group"},
			wantErr:         true,
			expectedValCall: 0,
			expectedRepCall: 0,
		},
		{
			name:            "nil group",
			ctx:             context.Background(),
			group:           nil,
			wantErr:         true,
			expectedValCall: 0,
			expectedRepCall: 0,
		},
		{
			name:            "validation error",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Test Group"},
			validReturns:    []any{errors.New("validation error")},
			wantErr:         true,
			expectedValCall: 1,
			expectedRepCall: 0,
		},
		{
			name:            "repo create error",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Test Group"},
			validReturns:    []any{nil},
			createReturns:   []any{errors.New("repo error")},
			wantErr:         true,
			expectedValCall: 1,
			expectedRepCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockGroupStorage(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.validReturns != nil {
				mockVal.On("ValidateGroup", mock.Anything, test.group).Return(test.validReturns...).Times(test.expectedValCall)
			}
			if test.createReturns != nil {
				mockRepo.On("Create", mock.Anything, test.group).Return(test.createReturns...).Times(test.expectedRepCall)
			}

			svc, err := NewGroupStorage(logger, mockRepo, mockVal, tracer)
			require.NoError(t, err, "unexpected error creating service")

			err = svc.Create(test.ctx, test.group)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			mockVal.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

// nolint: dupl
func TestGroupStorageUpdate(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name            string
		ctx             context.Context
		group           *entity.Group
		validReturns    []any
		updateReturns   []any
		wantErr         bool
		expectedValCall int
		expectedRepCall int
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Updated Group"},
			validReturns:    []any{nil},
			updateReturns:   []any{nil},
			wantErr:         false,
			expectedValCall: 1,
			expectedRepCall: 1,
		},
		{
			name:            "nil context",
			ctx:             nil,
			group:           &entity.Group{Id: "group1", Name: "Updated Group"},
			wantErr:         true,
			expectedValCall: 0,
			expectedRepCall: 0,
		},
		{
			name:            "nil group",
			ctx:             context.Background(),
			group:           nil,
			wantErr:         true,
			expectedValCall: 0,
			expectedRepCall: 0,
		},
		{
			name:            "validation error",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Updated Group"},
			validReturns:    []any{errors.New("validation error")},
			wantErr:         true,
			expectedValCall: 1,
			expectedRepCall: 0,
		},
		{
			name:            "repo update error",
			ctx:             context.Background(),
			group:           &entity.Group{Id: "group1", Name: "Updated Group"},
			validReturns:    []any{nil},
			updateReturns:   []any{errors.New("repo error")},
			wantErr:         true,
			expectedValCall: 1,
			expectedRepCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockGroupStorage(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.validReturns != nil {
				mockVal.On("ValidateGroup", mock.Anything, test.group).Return(test.validReturns...).Times(test.expectedValCall)
			}
			if test.updateReturns != nil {
				mockRepo.On("Update", mock.Anything, test.group).Return(test.updateReturns...).Times(test.expectedRepCall)
			}

			svc, err := NewGroupStorage(logger, mockRepo, mockVal, tracer)
			require.NoError(t, err, "unexpected error creating service")

			err = svc.Update(test.ctx, test.group)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			mockVal.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGroupStorageLoadAll(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name            string
		ctx             context.Context
		listAllReturns  []any
		wantErr         bool
		expectedRepCall int
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			listAllReturns:  []any{[]*entity.Group{{Id: "1"}, {Id: "2"}}, nil},
			wantErr:         false,
			expectedRepCall: 1,
		},
		{
			name:            "success empty list",
			ctx:             context.Background(),
			listAllReturns:  []any{[]*entity.Group{}, nil},
			wantErr:         false,
			expectedRepCall: 1,
		},
		{
			name:            "nil context",
			ctx:             nil,
			wantErr:         true,
			expectedRepCall: 0,
		},
		{
			name:            "repo error",
			ctx:             context.Background(),
			listAllReturns:  []any{nil, errors.New("repo error")},
			wantErr:         true,
			expectedRepCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockGroupStorage(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.listAllReturns != nil {
				mockRepo.On("ListAll", mock.Anything).Return(test.listAllReturns...).Times(test.expectedRepCall)
			}

			svc, err := NewGroupStorage(logger, mockRepo, mockVal, tracer)
			require.NoError(t, err, "unexpected error creating service")

			groups, err := svc.ListAll(test.ctx)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, groups, "groups should be nil on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, groups, "groups should not be nil")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGroupStorageLoad(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name            string
		ctx             context.Context
		id              string
		readReturns     []any
		wantErr         bool
		expectedRepCall int
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			id:              "group1",
			readReturns:     []any{&entity.Group{Id: "group1", Name: "Test Group"}, nil},
			wantErr:         false,
			expectedRepCall: 1,
		},
		{
			name:            "nil context",
			ctx:             nil,
			id:              "group1",
			wantErr:         true,
			expectedRepCall: 0,
		},
		{
			name:            "repo error",
			ctx:             context.Background(),
			id:              "group1",
			readReturns:     []any{nil, errors.New("repo error")},
			wantErr:         true,
			expectedRepCall: 1,
		},
		{
			name:            "group not found",
			ctx:             context.Background(),
			id:              "nonexistent",
			readReturns:     []any{nil, errors.New("not found")},
			wantErr:         true,
			expectedRepCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockGroupStorage(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.readReturns != nil {
				mockRepo.On("Read", mock.Anything, test.id).Return(test.readReturns...).Times(test.expectedRepCall)
			}

			svc, err := NewGroupStorage(logger, mockRepo, mockVal, tracer)
			require.NoError(t, err, "unexpected error creating service")

			group, err := svc.Load(test.ctx, test.id)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
				assert.Nil(t, group, "group should be nil on error")
			} else {
				assert.NoError(t, err, "expected no error")
				assert.NotNil(t, group, "group should not be nil")
				assert.Equal(t, test.id, group.Id)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGroupStorageRemove(t *testing.T) {
	logger := slog.Default()
	tracer := otel.Tracer("test")

	tests := []struct {
		name            string
		ctx             context.Context
		id              string
		deleteReturns   []any
		wantErr         bool
		expectedRepCall int
	}{
		{
			name:            "success",
			ctx:             context.Background(),
			id:              "group1",
			deleteReturns:   []any{nil},
			wantErr:         false,
			expectedRepCall: 1,
		},
		{
			name:            "nil context",
			ctx:             nil,
			id:              "group1",
			wantErr:         true,
			expectedRepCall: 0,
		},
		{
			name:            "repo delete error",
			ctx:             context.Background(),
			id:              "group1",
			deleteReturns:   []any{errors.New("repo error")},
			wantErr:         true,
			expectedRepCall: 1,
		},
		{
			name:            "group not found",
			ctx:             context.Background(),
			id:              "nonexistent",
			deleteReturns:   []any{errors.New("not found")},
			wantErr:         true,
			expectedRepCall: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRepo := mocks.NewMockGroupStorage(t)
			mockVal := servmocks.NewMockValidator(t)

			if test.deleteReturns != nil {
				mockRepo.On("Delete", mock.Anything, test.id).Return(test.deleteReturns...).Times(test.expectedRepCall)
			}

			svc, err := NewGroupStorage(logger, mockRepo, mockVal, tracer)
			require.NoError(t, err, "unexpected error creating service")

			err = svc.Remove(test.ctx, test.id)

			if test.wantErr {
				assert.Error(t, err, "expected an error")
			} else {
				assert.NoError(t, err, "expected no error")
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
