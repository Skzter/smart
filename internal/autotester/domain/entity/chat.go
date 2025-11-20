package entity

import (
	"time"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// Session represents a user session containing session summaries, requests, and responses.
type Chat struct {
	Id      string `json:"id"`
	UserId  string `json:"userId"`
	Created time.Time
	Updated time.Time

	Messages []entity.Message

	LastTest      string
	SystemPrompt  string
	InitialPrompt string
}
