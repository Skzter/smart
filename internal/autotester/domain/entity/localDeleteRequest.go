package entity

// LocalDeleteRequest represents a request to delete a locally stored test case.
type LocalDeleteRequest struct {
	TestcaseId     string `json:"testcaseId" form:"testcaseId"`
	UserId         string `json:"userId" form:"userId"`
	ConversationId string `json:"conversationId" form:"conversationId"`
}
