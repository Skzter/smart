package entity

// Session represents a user session containing session summaries, requests, and responses.
type Session struct {
	SessionId             string `json:"conversationId"`
	CurrentSessionSummary *SessionSummary
	History               []*SessionSummary
	UserRequests          []*UserRequest
	RequestsForLLM        []*RequestForLLM
	LlmResponses          []*LLMResponse
	ResponsesForUser      []*ResponseForUser
}

