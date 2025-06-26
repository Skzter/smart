package entity

// ResponseForUserDTO is a data transfer object for responses sent to the user.
// It contains the response text, session information, and log stamp.
type ResponseForUserDTO struct {
	ResponseText MessageDTO `json:"message"`
	SessionIdDTO
	LogStampDTO
}

// ToEntity converts the ResponseForUserDTO to a ResponseForUser entity.
// Returns an empty ResponseForUser.
func (ResponseForUserDTO) ToEntity() ResponseForUser {
	return ResponseForUser{}
}
