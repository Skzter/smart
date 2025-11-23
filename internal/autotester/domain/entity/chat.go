package entity

import (
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
