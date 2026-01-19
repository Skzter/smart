package entity

// UserChatSummaryQueryParameters is for binding the limit from the query in uri on the /users/:userId/chats endpoint
type UserChatSummaryQueryParameters struct {
	Page   int      `form:"page" binding:"omitempty,min=0"`
	Groups []string `form:"groups" binding:"omitempty,dive,required"`
}
