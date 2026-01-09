package entity

// UserChatSummaryParameters is for binding the userid in the uri on the /users/:userId/chats endpoint
type UserChatSummaryParameters struct {
	UserID string `uri:"userId" binding:"required,startswith=auth0"`
}

// UserChatSummaryQueryParameters is for binding the limit from the query in uri on the /users/:userId/chats endpoint
type UserChatSummaryQueryParameters struct {
	Limit int `form:"limit" binding:"omitempty,min=1"`
}
