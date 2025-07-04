package entity

// MessageDTO is a data transfer object for Message.
// It contains the message body and the actor information.
type MessageDTO struct {
	MessageBody string `json:"data"`
	Actor       string `json:"agent"`
}

// ToEntity converts the MessageDTO to a Message entity.
// Returns an empty Message.
func (Message) ToEntity() Message {
	return Message{}
}
