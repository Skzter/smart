package entity

// LocalDeleteRequest represents a request to delete a locally stored test case.
type LocalDeleteRequest struct {
	TestcaseId     string `json:"testcaseId"`
	UserId         string `json:"userId"`
	ConversationId string `json:"conversationId"`
}
