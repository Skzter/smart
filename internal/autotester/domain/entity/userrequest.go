package entity

type UserRequest struct {
	SessionId
	logStamp   LogStamp
	userPrompt *UserPrompt
}
