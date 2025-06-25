package entity

type ResponseForUserDTO struct {
	ResponseText Message `json:"message"`
	SessionId
	LogStampDTO
}

func (ResponseForUserDTO) ToEntity() ResponseForUser {
	return ResponseForUser{}
}
