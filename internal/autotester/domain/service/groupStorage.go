package service

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GroupStorage provides an interface to persist Group entities.
type GroupStorage interface {
	// Create persists a new Group entity into the storage.
	// Returns an error if the operation fails.
	Create(ctx context.Context, group *entity.Group) error
	// Update updates an existing Group entity in the storage.
	// Returns an error if the operation fails.
	Update(ctx context.Context, group *entity.Group) error
	// ListAll retrieves all Group entities from storage.
	// Returns a slice of groups or an error if the operation fails.
	ListAll(ctx context.Context) ([]*entity.Group, error)
	// Load retrieves a Group entity from storage by its id.
	// Returns the group or an error if the operation fails.
	Load(ctx context.Context, id string) (*entity.Group, error)
	// Remove deletes a Group entity from storage by its id.
	// Returns an error if the operation fails.
	Remove(ctx context.Context, id string) error
}

// groupStorage implements the GroupStorage interface
// and provides logic for storing Group entities via the underlying repository.
type groupStorage struct {
	logger    *slog.Logger
	repo      repository.GroupStorage
	validator Validator
	tracer    trace.Tracer
}

// NewGroupStorage creates a new GroupStorage instance.
// Returns the service or an error if any of the arguments are nil.
func NewGroupStorage(logger *slog.Logger, repo repository.GroupStorage, validator Validator, tracer trace.Tracer) (GroupStorage, error) {
	if err := assert.NotNil(logger, repo, validator, tracer); err != nil {
		return nil, err
	}

	return &groupStorage{
		logger:    logger,
		repo:      repo,
		validator: validator,
		tracer:    tracer,
	}, nil
}

// Create persists a new Group entity into the storage.
func (s *groupStorage) Create(ctx context.Context, group *entity.Group) error {
	if err := assert.NotNil(ctx, group); err != nil {
		return err
	}
	ctx, span := s.tracer.Start(ctx, "groupStorage.Create")
	defer span.End()

	if err := s.validator.ValidateGroup(ctx, group); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error during validation")
		return err
	}

	if err := s.repo.Create(ctx, group); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while creating group")
		return err
	}

	s.logger.Debug("group successfully created",
		slog.String("groupId", group.Id),
	)
	span.SetStatus(codes.Ok, "")
	return nil
}

// Update updates an existing Group entity in the storage.
func (s *groupStorage) Update(ctx context.Context, group *entity.Group) error {
	if err := assert.NotNil(ctx, group); err != nil {
		return err
	}
	ctx, span := s.tracer.Start(ctx, "groupStorage.Update")
	defer span.End()

	if err := s.validator.ValidateGroup(ctx, group); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error during validation")
		return err
	}

	if err := s.repo.Update(ctx, group); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while updating group")
		return err
	}

	s.logger.Debug("group successfully updated",
		slog.String("groupId", group.Id),
	)
	span.SetStatus(codes.Ok, "")
	return nil
}

// ListAll retrieves all Group entities from storage.
func (s *groupStorage) ListAll(ctx context.Context) ([]*entity.Group, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "groupStorage.ListAll")
	defer span.End()

	groups, err := s.repo.ListAll(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while loading all groups")
		return nil, err
	}

	s.logger.Debug("groups successfully loaded",
		slog.Int("count", len(groups)),
	)
	span.SetStatus(codes.Ok, "")
	return groups, nil
}

// Load retrieves a Group entity from storage by its id.
func (s *groupStorage) Load(ctx context.Context, id string) (*entity.Group, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}
	ctx, span := s.tracer.Start(ctx, "groupStorage.Load")
	defer span.End()

	group, err := s.repo.Read(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while loading group")
		return nil, err
	}

	s.logger.Debug("group successfully loaded",
		slog.String("groupId", id),
	)
	span.SetStatus(codes.Ok, "")
	return group, nil
}

// Remove deletes a Group entity from storage by its id.
func (s *groupStorage) Remove(ctx context.Context, id string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	ctx, span := s.tracer.Start(ctx, "groupStorage.Remove")
	defer span.End()

	if err := s.repo.Delete(ctx, id); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "error while deleting group")
		return err
	}

	s.logger.Debug("group successfully deleted",
		slog.String("groupId", id),
	)
	span.SetStatus(codes.Ok, "")
	return nil
}
