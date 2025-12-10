package entity

// RunTestOutput represents the result of a test run.
// Success indicates whether the run completed successfully.
// Result holds the test output when successful.
// Error holds a human-readable error message when not successful.
type RunTestOutput struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}
