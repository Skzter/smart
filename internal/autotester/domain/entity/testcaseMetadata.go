package entity

// TestcaseMetadata represents metadata information for a stored test case.
type TestcaseMetadata struct {
	Key        string `json:"key"`
	TestcaseId string `json:"testcaseId"`
	Author     string `json:"author"`
	Created    string `json:"created"`
	Updated    string `json:"updated"`
	Name       string `json:"name"`
}
