package service

// AutotesterAPIService provides business logic for interacting with the Autotester API.
type AutotesterAPIService interface {
	// GetTemplate retrieves the test generation template.
	GetTemplate()

	// GenerateTest creates a new test from the provided specification.
	GenerateTest()

	// ExecuteTest runs an existing test by ID.
	ExecuteTest()
}
