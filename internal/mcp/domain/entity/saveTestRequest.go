package entity

// SaveTestRequest represents a request to persist a generated test
// locally or in a storage backend. Extend with fields like TestID,
// Code, UserID when persistence is implemented.
type SaveTestRequest struct {
	Prompt         string `json:"prompt"`
	UserId         string `json:"userId"`
	ConversationId string `json:"conversationId"`
}
