package entity

// SessionId represents the unique identifier for a session.
type SessionId struct {
	Id string
}

// ToDTO converts the SessionId to a SessionDTO.
// Returns an empty SessionDTO.
func (SessionId) ToDTO() SessionDTO {
	return SessionDTO{}
}
