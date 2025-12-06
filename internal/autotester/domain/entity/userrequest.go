package entity

// UserRequest represents a user request within a session, including prompt and log information.
type UserRequest struct {
	ChatId string `json:"conversationId"`
	Prompt string `json:"prompt"`
	UserId string `json:"userId"`
}
