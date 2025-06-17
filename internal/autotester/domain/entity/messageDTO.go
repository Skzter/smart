package entity

type MessageDTO struct {
	MessageBody string `json:"message"`
	Actor       string `json:"agent"`
}

func (Message) ToEntity() Message {
	return Message{}
}
