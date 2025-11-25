package entity

import (
	"time"

	// SessionSummary represents a summary of an autotester session including user messages.
	// This resurrects a previously referenced type that tests expect.
	// Keep fields minimal based on test usage patterns (Summary, CreatedAt, Messages).
	// Message type is reused from shared domain entity.

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// SessionSummary holds a textual summary, creation timestamp and a slice of messages.
// Messages may be empty but not nil for validation routines.
// NOTE: If more fields were present historically, extend cautiously to satisfy repository logic.
//
//nolint:lll
type SessionSummary struct {
	Summary   string            `json:"summary"`
	CreatedAt time.Time         `json:"createdAt"`
	Messages  []*shared.Message `json:"messages"`
}
