package entity

import (
	"errors"
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// Chat represents a single Chat, identified by a unique id and associated with a user.
type Chat struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Title    string           `json:"title"`
	Messages []entity.Message `json:"messages"`

	LastTest      string `json:"lastTest"`
	SystemPrompt  string `json:"systemPrompt"`
	InitialPrompt string `json:"initialPrompt"`
}

// Validate validates a Chat entity.
// Returns an error if any required field is empty or invalid.
func (chat *Chat) Validate() error {
	switch {
	case chat.InitialPrompt == "":
		return errors.New("initial prompt is empty")
	case chat.SystemPrompt == "":
		return errors.New("system prompt is empty")
	case chat.Id == "":
		return errors.New("id is empty")
	case chat.UserId == "":
		return errors.New("userId is empty")
	case len(chat.Messages) == 0:
		return errors.New("contains no messages")
	}

	for _, msg := range chat.Messages {
		switch {
		case msg.Body == "":
			return errors.New("contains empty messages")
		case msg.Id == "":
			return errors.New("contains messages with empty id")
		case msg.Role == "":
			return errors.New("contains message with empty role")
		}
	}
	return nil
}
