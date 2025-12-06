package entity

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a message entity with an actor and message body.
type Message struct {
	Id        string    `json:"id"`
	Role      string    `json:"role"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

// NewMessage creates a new Message with the given the given role and the given user
func NewMessage(body string, role string) *Message {
	return &Message{
		Id:        uuid.NewString(),
		Role:      role,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
}
