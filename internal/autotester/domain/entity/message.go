package entity

// Message represents a message entity with an actor and message body.
type Message struct {
	Actor       string
	MessageBody string
}

// ToDTO converts the Message to a MessageDTO.
// Returns an empty MessageDTO.
func (Message) ToDTO() MessageDTO {
	return MessageDTO{}
}
