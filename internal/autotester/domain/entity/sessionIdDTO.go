package entity

type SessionIdDTO struct {
	Id string `json:"conversationId"`
}

func (SessionIdDTO) ToEntity() SessionId {
	return SessionId{}
}
