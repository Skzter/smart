package entity

// RunTestRequest is the entity for running the test with all IDs
type RunTestRequest struct {
	UserID string `json:"userId"`
	TestId string `json:"testId"`
	ChatID string `json:"chatId"`
}
