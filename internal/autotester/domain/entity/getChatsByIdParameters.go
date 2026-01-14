package entity

// UserChatsParameter is for binding the chatid from the uri on the /chats/:chatId endpoint
type UserChatsParameter struct {
	ChatID string `uri:"chatId" binding:"required,uuid"`
}
