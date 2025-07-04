package entity

import "time"

// SessionSummary represents a summary of a session, including all messages and the creation time.
type SessionSummary struct {
	Summary   string // concatenation of messages in the correct order
	CreatedAt time.Time
	Messages  []*Message
}
