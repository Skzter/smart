package entity

// ValidationResult represents the normalized outcome of a JWT validation
type ValidationResult struct {
	Valid   bool
	Revoked bool
}
