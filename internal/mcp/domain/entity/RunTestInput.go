package entity

// RunTestInput is the input payload for executing a test.
// UserID is the identifier of the user requesting the run.
// TestId is the identifier of the test to run.
// SessionID is the identifier of the session.
type RunTestInput struct {
	UserID    string `json:"userId"`
	TestId    string `json:"testId"`
	SessionID string `json:"sessionId"`
}
