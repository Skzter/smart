package entity

type Message struct {
	Actor       string
	MessageBody string
}

func (Message) ToDTO() MessageDTO {
	return MessageDTO{}
}
