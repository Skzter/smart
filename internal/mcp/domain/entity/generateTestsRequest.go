package entity

// GenerateTestRequest contains the information required by the
// Autotester to create a new test (prompt, user and chat ids).
type GenerateTestRequest struct {
	Prompt string `json:"prompt"`
	UserId string `json:"userId"`
	ChatId string `json:"chatId"`
}
