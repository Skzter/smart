package entity

// RunTestRequest represents a request to run a previously stored test.
// Fields can be added later (e.g. TestID, Options) when needed.
type RunTestRequest struct {
	Prompt         string `json:"prompt"`
	UserId         string `json:"userId"`
	ConversationId string `json:"conversationId"`
}
