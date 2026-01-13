package entity

// RemoveChatFromGroupRequest represents the request parameters for removing a chat from a group
type RemoveChatFromGroupRequest struct {
	GroupId string `uri:"groupId" binding:"required,uuid"`
	ChatId  string `uri:"chatId" binding:"required,uuid"`
}
