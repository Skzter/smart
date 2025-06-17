package entity

type UserRequest struct {
	SessionId
	LogStamp   LogStamp
	UserPrompt *UserPrompt
}

func (UserRequest) ToDTO() UserRequestDTO {
	return UserRequestDTO{}
}
