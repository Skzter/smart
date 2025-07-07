package entity

// Message represents a message entity with an actor and message body.
type Message struct {
	Actor       string `json:"agent"`
	MessageBody string `json:"data"`
}
