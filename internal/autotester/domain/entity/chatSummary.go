package entity

import (
	"strings"
	"time"
)

// ChatSummary represents metadata of a Chat.
type ChatSummary struct {
	ChatId string   `json:"chatId"`
	Author string   `json:"userId"`
	Groups []string `json:"groups"`

	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	LastModifiedBy string    `json:"lastModifiedBy"`
}

// Cmp compares two ChatSummary instances, ordering by UpdatedAt descending, then by ChatId ascending.
// Returns negative if a < b, zero if a == b, positive if a > b.
func (a *ChatSummary) Cmp(b *ChatSummary) int {
	if updated := -a.UpdatedAt.Compare(b.UpdatedAt); updated != 0 {
		return updated
	}
	return strings.Compare(a.ChatId, b.ChatId)
}
