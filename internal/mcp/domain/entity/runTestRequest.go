package entity

// RunTestRequest represents a request to run a previously stored test.
// Fields can be added later (e.g. TestID, Options) when needed.
type RunTestRequest struct {
	TestId string `json:"testId"`
	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
}
