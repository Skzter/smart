package entity

// ValidationResult represents the normalized outcome of a DB-backed bearer token validation
type ValidationResult struct {
	Valid   bool
	Revoked bool
}
