package entity

import "time"

// Message represents a message entity with an actor and message body.
type Message struct {
	Id        string    `json:"id"`
	Role      string    `json:"agent"`
	Body      string    `json:"data"`
	CreatedAt time.Time `json:"createdAt"`
}
