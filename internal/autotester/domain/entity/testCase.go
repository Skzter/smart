package entity

// TestCase represents a single test case with its ID, description, expected output, code, and status.
type TestCase struct {
	TestID      string
	Description string
	TestCode    TestCode
	Status      TestStatus
}

// TestCode contains the code to be executed for a test case.
type TestCode struct {
	Code string
}
