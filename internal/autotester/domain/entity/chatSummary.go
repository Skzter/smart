package entity

// Chat represents a chat entity with a title and an assigned user.
type ChatSummary struct {
	ChatId       string
	UserId       string
	Title        string
	MessageCount int
}
