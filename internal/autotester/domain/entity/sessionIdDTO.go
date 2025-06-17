package entity

type SessionIdDTO struct {
	Id string `json:"ConversationId"`
}

func (SessionIdDTO) ToEntity() SessionId {
	return SessionId{}
}
