package entity

// ReadTestLogStreamIn represents the input parameters for the ReadTestLogStream tool.
// It specifies which test's logs to read and optionally from which point (cursor) to continue.
type ReadTestLogStreamIn struct {
	TestID string `json:"testId"`
	Cursor string `json:"cursor,omitempty"`
}
