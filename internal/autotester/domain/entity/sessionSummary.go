package entity

import "time"

type SessionSummary struct {
	summary   string
	createdAt time.Time
	messages  []*Message
}
