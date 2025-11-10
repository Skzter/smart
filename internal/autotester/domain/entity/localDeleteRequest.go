package entity

// LocalDeleteRequest represents a request to delete a locally stored test case.
type LocalDeleteRequest struct {
	TestcaseId     string `form:"testcaseId"`
	UserId         string `form:"userId"`
	ConversationId string `form:"conversationId"`
}
