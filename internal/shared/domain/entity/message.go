package entity

import "time"

// Message represents a message entity with an actor and message body.
type Message struct {
	Id        string    `json:"id"`
	Role      string    `json:"role"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}
