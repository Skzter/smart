package entity

type ResponseForUserDTO struct {
	ResponseText MessageDTO `json:"message"`
	SessionIdDTO
	LogStampDTO
}

func (ResponseForUserDTO) ToEntity() ResponseForUser {
	return ResponseForUser{}
}
