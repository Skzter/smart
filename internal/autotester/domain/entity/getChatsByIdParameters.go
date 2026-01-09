package entity

// UserChatsParameter is for binding the userid and chatid from the uri on the /users/:userId/chats/:chatId endpoint
type UserChatsParameter struct {
	UserID string `uri:"userId" binding:"required,startswith=auth0"`
	ChatID string `uri:"chatId" binding:"required,uuid"`
}
