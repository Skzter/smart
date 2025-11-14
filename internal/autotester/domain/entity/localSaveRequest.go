package entity

// LocalSaveRequest represents a request to save a test locally.
type LocalSaveRequest struct {
	UserId         string `json:"userId"`
	ConversationId string `json:"conversationId"`
	Code           string `json:"code"`
}
