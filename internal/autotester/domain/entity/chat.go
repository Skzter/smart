package entity

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Type represents a Type of Message stored in Chat entity
type Type uint8

const (
	// TypeValidation is the Type of messages used in validation
	TypeValidation = 1 << iota
	// TypeGeneration is the Type of messages used in generation
	TypeGeneration
	// TypeAny matches any type, and is mainly used for the users messages
	TypeAny = Type(math.MaxUint8)
)

func types() []Type {
	return []Type{TypeAny, TypeGeneration, TypeValidation}
}

// Chat represents a single Chat, identified by a unique id and associated with a user.
type Chat struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title     string    `json:"title"`

	Messages []Message `json:"messages"`
	index    map[Type][]int

	LastTest                 string `json:"lastTest"`
	LastAutoPlaywrightPrompt string `json:"lastAutoPlaywrightPrompt"`
}

// Validate validates a Chat entity.
// Returns an error if any required field is empty or invalid.
func (chat *Chat) Validate() error {
	if err := assert.StringsNotEmpty(
		chat.Id,
		chat.UserId); err != nil {
		return err
	}

	if chat.UpdatedAt.IsZero() {
		return errors.New("updatedAt cannot be zero")
	}
	if chat.CreatedAt.IsZero() {
		return errors.New("createdAt cannot be zero")
	}

	if err := assert.ArrayLengthGreaterThan(chat.Messages, 0); err != nil {
		return err
	}

	// shared validation for each Message in Messages
	return nil
}

// NewChat creates a new chat with for the given user with the given messages
func NewChat(userId string, messages []Message) *Chat {
	now := time.Now().UTC()
	return &Chat{
		Id:        uuid.NewString(),
		UserId:    userId,
		CreatedAt: now,
		UpdatedAt: now,
		Messages:  messages,

		index: nil,
	}
}

// AddMessage adds a Message of the given type to the chats Messages
func (m *Chat) AddMessage(message *shared.Message, ts ...Type) {
	t := TypeAny
	if len(ts) > 0 {
		t = ts[0]
	}
	m.Messages = append(m.Messages, Message{Message: *message, Type: t})
	m.index = nil // index muss neu erstellt werden (lazy reicht aus)
}

func (m *Chat) buildIndex() {
	if m.index != nil {
		return
	}
	m.index = make(map[Type][]int)

	for i, msg := range m.Messages {
		for _, t := range types() {
			if msg.Type&t > 0 {
				m.index[t] = append(m.index[t], i)
			}
		}
	}
}

// Filter returns a slice of all Messages associated with the Type
func (m *Chat) Filter(t Type) []*shared.Message {
	m.buildIndex()

	index, ok := m.index[t]
	if !ok {
		return nil
	}

	result := make([]*shared.Message, 0, len(m.Messages)) // pre allocation for memory efficiency

	for _, i := range index {
		msgcopy := m.Messages[i].Message
		result = append(result, &msgcopy)
	}

	return result
}

// CountMessages returns the number of Messages associated with the Type
func (m *Chat) CountMessages(t Type) int {
	m.buildIndex()
	index, ok := m.index[t]
	if !ok {
		return 0
	}
	return len(index)
}
