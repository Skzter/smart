package entity

// UserRequestDTO is a data transfer object for a user request.
// It contains the message, user ID, and session ID.
type UserRequestDTO struct {
	Message   MessageDTO `json:"message"`
	UserID    string     `json:"userId"`
	SessionId string     `json:"conversationId"`
}

// ToEntity converts the UserRequestDTO to a UserRequest entity.
// Returns an empty UserRequest.
func (u UserRequestDTO) ToEntity() UserRequest {
	return UserRequest{}
}
