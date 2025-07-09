package entity

// UserRequest represents a user request within a session, including prompt and log information.
type UserRequest struct {
	SessionId  string   `json:"conversationId"`
	LogStamp   LogStamp `json:"-"`
	UserPrompt *UserPrompt
	Message    Message `json:"message"`
	UserId     string  `json:"userId"`
}
