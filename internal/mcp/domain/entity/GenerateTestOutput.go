package entity

// GenerateTestOutput represents the result of a test generation request.
// TestID is the identifier of the created test.
// TestCode contains the generated test source code.
// TestStatus indicates the current state of the generated test (e.g., "created", "failed").
type GenerateTestOutput struct {
	TestID     string `json:"test_id"`
	TestCode   string `json:"test_code"`
	TestStatus string `json:"test_status"`
}
