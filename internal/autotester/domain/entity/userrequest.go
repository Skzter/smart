package entity

type UserRequest struct {
	SessionId  string   `json:"conversationId"`
	LogStamp   LogStamp `json:"-"`
	UserPrompt *UserPrompt
	Message    Message `json:"message"`
	UserId     string  `json:"userId"`
}
