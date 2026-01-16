package entity

// UpdateChatUri represents the URI parameters required to update a chat.
type UpdateChatUri struct {
	UserId string `uri:"userId" binding:"required"`
	ChatId string `uri:"chatId" binding:"required"`
}
