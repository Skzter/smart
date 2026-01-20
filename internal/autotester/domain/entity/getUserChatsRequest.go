package entity

// UserChatSummaryQueryParameters is for binding the limit from the query in uri on the /users/:userId/chats endpoint
type UserChatSummaryQueryParameters struct {
	Limit  int      `form:"limit" binding:"omitempty,min=1"`
	Groups []string `form:"groups" binding:"omitempty,dive,required"`
}
