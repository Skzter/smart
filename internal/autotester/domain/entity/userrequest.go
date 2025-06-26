package entity

// UserRequest represents a user request within a session, including prompt and log information.
type UserRequest struct {
	SessionId
	LogStamp   LogStamp
	UserPrompt *UserPrompt
}

// ToDTO converts the UserRequest to a UserRequestDTO.
// Returns an empty UserRequestDTO.
func (UserRequest) ToDTO() UserRequestDTO {
	return UserRequestDTO{}
}
