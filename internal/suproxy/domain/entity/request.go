package entity

import (
	"time"
)

type Request struct {
	RequestID int       `json:"request_id"`
	PromptID  int       `json:"prompt_id"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}
