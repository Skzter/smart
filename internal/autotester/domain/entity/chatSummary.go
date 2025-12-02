package entity

import "time"

// ChatSummary represents metadata of a Chat.
type ChatSummary struct {
	ChatId    string    `json:"chatId"`
	UserId    string    `json:"userId"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
