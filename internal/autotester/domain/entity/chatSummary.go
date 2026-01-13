package entity

import "time"

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
