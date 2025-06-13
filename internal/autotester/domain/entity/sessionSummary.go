package entity

import "time"

type SessionSummary struct {
	Summary   string // concatenation of messages in the correct order
	CreatedAt time.Time
	Messages  []*Message
}
