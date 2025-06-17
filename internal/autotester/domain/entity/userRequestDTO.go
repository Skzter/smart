package entity

type UserRequestDTO struct {
	Message   MessageDTO `json:"message"`
	UserID    string     `json:"userId"`
	SessionId string     `json:"conversationId"`
}

func (u UserRequestDTO) ToEntity() UserRequest {
	return UserRequest{}
}
