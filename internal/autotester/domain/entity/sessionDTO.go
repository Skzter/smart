package entity

type SessionDTO struct {
	SessionId string    `json:"ConversationId"`
	Messages  []Message `json:"messages"`
}

func (SessionDTO) ToEntity() Session {
	return Session{}
}
