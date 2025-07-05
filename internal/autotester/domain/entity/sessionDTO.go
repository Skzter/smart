package entity

// SessionDTO is a data transfer object for a user session.
// It contains the session ID and a list of messages.
type SessionDTO struct {
	SessionId `json:"ConversationId"`
	Messages  []Message `json:"messages"`
}

// ToEntity converts the SessionDTO to a Session entity.
// Returns an empty Session.
func (SessionDTO) ToEntity() Session {
	return Session{}
}
