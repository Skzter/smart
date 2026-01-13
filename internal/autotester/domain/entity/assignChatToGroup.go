package entity

// AssignChatToGroupRequestURI represents the URI parameters for assigning a chat to a group
type AssignChatToGroupRequestURI struct {
	ChatId string `uri:"chatId" binding:"required,uuid"`
}

// AssignChatToGroupRequestJSON represents the JSON payload for assigning a chat to a group
type AssignChatToGroupRequestJSON struct {
	GroupIds []string `json:"groupIds" binding:"min=1,unique,dive,required,uuid"`
}
