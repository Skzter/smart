package entity

// UserRequest represents a user request within a chat, including prompt and log information.
type UserRequest struct {
	ChatId string `json:"chatId"`
	Prompt string `json:"prompt"`
	UserId string `json:"userId"`
}
