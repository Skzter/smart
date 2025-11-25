package entity

import (
	"time"

	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Chat represents a single Chat, identified by a unique id and associated with a user.
type Chat struct {
	Id        string    `json:"id"`
	UserId    string    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Title    string           `json:"title"`
	Messages []entity.Message `json:"messages"`

	LastTest                 string `json:"lastTest"`
	LastAutoPlaywrightPrompt string `json:"lastAutoPlaywrightPrompt"`
	LastValidationPrompt     string `json:"lastValidationPrompt"`
	InitialUserPrompt        string `json:"initiaUserlPrompt"`
}

// AddMessage adds a message to the chats Messages with the provied role and body
func (c *Chat) AddMessage(body string, role string) {
	c.Messages = append(c.Messages, entity.Message{
		Id:        uuid.NewString(),
		Role:      role,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	})
}

// Validate validates a Chat entity.
// Returns an error if any required field is empty or invalid.
func (chat *Chat) Validate() error {
	if err := assert.StringsNotEmpty(
		chat.Id,
		chat.UserId,
		chat.LastAutoPlaywrightPrompt,
		chat.LastValidationPrompt,
		chat.InitialUserPrompt); err != nil {
		return err
	}

	if err := assert.ArrayLengthGreaterThan(chat.Messages, 0); err != nil {
		return err
	}

	for _, msg := range chat.Messages {
		if err := assert.StringsNotEmpty(msg.Id, msg.Role, msg.Body); err != nil {
			return err
		}
	}
	return nil
}
