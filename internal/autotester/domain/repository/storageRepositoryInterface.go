package repository

import "context"

// StorageRepository defines the interface for persistent data storage operations.
// It provides basic CRUD functionality (Create, Read, Update, Delete) for abstracted storage systems
type StorageRepository[T any] interface {

	// Create stores a new record or object in the underlying storage system.
	Create(ctx context.Context, obj *T) (string, error)

	// Read retrieves a record or object from the storage based on an identifier or query.
	Read(key string) (*T, error) // return data or error, Johannes fragen

	// Update modifies an existing record or object in the storage system.
	Update(ctx context.Context, obj *T, key string) (string, error)

	// Delete removes a record or object from the storage system.
	Delete(key string) error
}
