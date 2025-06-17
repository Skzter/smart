package entity

type UserRequest struct {
	SessionId
	LogStamp   LogStamp
	UserPrompt *UserPrompt
}
