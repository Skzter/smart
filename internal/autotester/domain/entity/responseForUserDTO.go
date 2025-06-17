package entity

type ResponseForUserDTO struct {
	SessionId    `json:"ConversationId"`
	LogStamp     LogStampDTO
	ResponseText ModelAnswerTextDTO
}

func (ResponseForUserDTO) ToEntity() ResponseForUser {
	return ResponseForUser{}
}
