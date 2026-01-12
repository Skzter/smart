package entity

// CreateGroupRequest represents the payload for creating a new group
type CreateGroupRequest struct {
	UserId      string `required,json:"userId"`
	GroupName   string `required,json:"groupName"`
	Description string `json:"description"`
}

// CreateGroupResponse represents the response after successfully creating a group
type CreateGroupResponse struct {
	GroupId string `json:"groupId"`
}

// AssignChatToGroupRequestURI represents the URI parameters for assigning a chat to a group
type AssignChatToGroupRequestURI struct {
	ChatId string `uri:"chatId" binding:"required,uuid"`
}

// AssignChatToGroupRequestJSON represents the JSON payload for assigning a chat to a group
type AssignChatToGroupRequestJSON struct {
	GroupId string `json:"groupId" binding:"required,uuid"`
}

// RemoveChatFromGroupRequest represents the request parameters for removing a chat from a group
type RemoveChatFromGroupRequest struct {
	GroupId string `uri:"groupId" binding:"required,uuid"`
	ChatId  string `uri:"chatId" binding:"required,uuid"`
}
