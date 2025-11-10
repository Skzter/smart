package entity

// LocalSaveResponse represents the response after saving a test case locally.
type LocalSaveResponse struct {
	TestcaseId string `json:"testcaseId"`
	Action     string `json:"action"`
}
