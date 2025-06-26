package repository

// StorageRepository defines the interface for persistent data storage operations.
// It provides basic CRUD functionality (Create, Read, Update, Delete) for abstracted storage systems
type StorageRepository interface {
	// TODO: Add parameter types and return types to methods

	// Create stores a new record or object in the underlying storage system.
	Create() // wegen key Johannes fragen

	// Frage:
	// wegen datenreingabe für 'Create()' und 'Update', Interface für Objekte erstellen die reingegeben werden können?
	// oder wie abstrahieren was alles gespeichert werden kann

	// Read retrieves a record or object from the storage based on an identifier or query.
	Read(key string) // return data or error, Johannes fragen

	// Update modifies an existing record or object in the storage system.
	Update(key string) error

	// Delete removes a record or object from the storage system.
	Delete(key string) error
}
