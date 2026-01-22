package entity

// ValidatePromptResponse represents the response payload returned by the
// Autotester/chat backend when a prompt validation was performed.
type ValidatePromptResponse struct {
	Result ValidateMessage `json:"message"`
	UserId string          `json:"userId"`
	ChatId string          `json:"chatId"`
}
