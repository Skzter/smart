package entity

// UpdateChatUri represents the URI parameters required to update a chat.
type UpdateChatUri struct {
	ChatId string `uri:"chatId" binding:"required"`
}
