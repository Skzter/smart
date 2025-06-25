package entity

type UserRequestDTO struct {
	Message   Message `json:"message"`
	UserID    string  `json:"userId"`
	SessionId string  `json:"conversationId"`
}

func (u UserRequestDTO) ToEntity() UserRequest {
	return UserRequest{}
}
