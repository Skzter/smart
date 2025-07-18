package entity

// TestStatus defines the status of a test as a string type.
type TestStatus string

// TestStatusPending, TestStatusRunning, TestStatusPassed, TestStatusFailed, and TestStatusSkipped
// are the possible values for TestStatus.
const (
	TestStatusPending TestStatus = "pending"
	TestStatusRunning TestStatus = "running"
	TestStatusPassed  TestStatus = "passed"
	TestStatusFailed  TestStatus = "failed"
	TestStatusSkipped TestStatus = "skipped"
)
