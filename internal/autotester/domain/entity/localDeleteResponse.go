package entity

// LocalDeleteResponse represents the response after deleting a test case locally.
type LocalDeleteResponse struct {
	TestcaseId string `json:"testcaseId"`
	Action     string `json:"action"`
}
