package entity

import (
	"time"

	"github.com/google/uuid"

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// MessageType represents a Type of Message stored in Chat entity
// ENUM(Validation, Generation, User)
type MessageType uint

//go:generate go tool go-enum -f=$GOFILE --marshal

// Chat represents a single Chat, identified by a unique id and associated with a user.
type Chat struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title     string    `json:"title"`

	Messages []*Message `json:"messages"`

	LastTest                 string `json:"lastTest"`
	LastAutoPlaywrightPrompt string `json:"lastAutoPlaywrightPrompt"`
}

// NewChat creates a new chat with for the given user with the given messages
func NewChat(userId string, messages []*Message) *Chat {
	now := time.Now().UTC()
	return &Chat{
		Id:        uuid.NewString(),
		UserId:    userId,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  messages,
	}
}

// AddMessage adds a Message of the given type to the chats Messages
func (m *Chat) AddMessage(message *shared.Message, ts ...MessageType) {
	t := MessageTypeUser
	if len(ts) > 0 {
		t = ts[0]
	}
	m.Messages = append(m.Messages, &Message{Message: *message, Type: t})
}

func (m *Chat) buildIndex() map[MessageType][]int {
	index := make(map[MessageType][]int)

	for i, msg := range m.Messages {
		if msg.Type == MessageTypeUser {
			for t := range _MessageTypeMap {
				index[t] = append(index[t], i)
			}
		} else {
			index[msg.Type] = append(index[msg.Type], i)
		}
	}

	return index
}

// Filter returns a slice of all Messages associated with the Type
func (m *Chat) Filter(t MessageType) []*shared.Message {
	indeces := m.buildIndex()

	index, ok := indeces[t]
	if !ok {
		return nil
	}

	result := make([]*shared.Message, 0, len(index)) // pre allocation for memory efficiency

	for _, i := range index {
		result = append(result, &m.Messages[i].Message)
	}

	return result
}
