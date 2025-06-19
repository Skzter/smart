package entity

type SessionId struct {
	Id string
}

func (SessionId) ToDTO() SessionDTO {
	return SessionDTO{}
}
