package entity

type MessageDTO struct {
	MessageBody string `json:"data"`
	Actor       string `json:"agent"`
}

func (Message) ToEntity() Message {
	return Message{}
}
