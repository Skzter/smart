package repository

// AutotesterAPIRepository handles HTTP communication with the Autotester backend API.
type AutotesterAPIRepository interface {
	// GetTemplate fetches the test generation template from the API.
	GetTemplate()

	// GenerateTest sends a test generation request to the API.
	GenerateTest()

	// SaveTest persists a generated test to the API.
	SaveTest()

	// RunTest executes a test by ID via the API.
	RunTest()
}
