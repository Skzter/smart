package entity

import "time"

// GenerateMessage represents a single message that the MCP/Autotester API
// can return as part of the response when a test is generated from a prompt.
type GenerateMessage struct {
	Id        string    `json:"id"`
	Role      string    `json:"role"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}
