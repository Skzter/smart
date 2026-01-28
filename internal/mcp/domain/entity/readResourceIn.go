package entity

// ReadTestLogStreamIn represents the input parameters for the ReadTestLogs tool.
// It specifies which test's logs to read.
type ReadTestLogStreamIn struct {
	TestId string `json:"testId"`
}
