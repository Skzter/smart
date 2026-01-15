package entity

// ReadTestLogStreamIn represents the input parameters for the ReadTestLogStream tool.
// It specifies which test's logs to read.
type ReadTestLogStreamIn struct {
	TestId string `json:"testId"`
}
