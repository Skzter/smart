package entity

// SessionIdDTO is a data transfer object for SessionId.
// It contains the unique session identifier.
type SessionIdDTO struct {
	Id string `json:"conversationId"`
}

// ToEntity converts the SessionIdDTO to a SessionId entity.
// Returns an empty SessionId.
func (SessionIdDTO) ToEntity() SessionId {
	return SessionId{}
}
