package repository

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// GroupStorageRepository defines the interface for a repository managing Group entities.
type GroupStorageRepository interface {
	// Create stores the provided Group object in the underlying storage system.
	// The storage key is generated from the group's Id, so duplicate entities will be overwritten.
	Create(ctx context.Context, obj *entity.Group) error

	// Read retrieves a Group object from storage by its groupId.
	Read(ctx context.Context, groupId string) (*entity.Group, error)

	// Delete removes a Group object from storage by its groupId.
	Delete(ctx context.Context, groupId string) error

	// ListAll retrieves all Groups from storage.
	ListAll(ctx context.Context) ([]*entity.Group, error)

	// Update updates an existing Group object in storage.
	Update(ctx context.Context, obj *entity.Group) error
}
