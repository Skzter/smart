package entity

import (
	"time"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// SessionSummary represents a summary of a session, including all messages and the creation time.
type SessionSummary struct {
	Summary   string // concatenation of messages in the correct order
	CreatedAt time.Time
	Messages  []*sharedEntity.Message
}
