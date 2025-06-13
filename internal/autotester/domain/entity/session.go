package entity

// Session represents a conversation session, holding communication elements and the history of the conversation
type Session struct {
	SessionId             // unique identifier of the session, embedded struct
	CurrentSessionSummary SessionSummary
	History               []*SessionSummary
	UserRequests          []*UserRequest
	RequestsForLLM        []*RequestForLLM
	LlmResponses          []*LLMResponse
	ResponsesForUser      []*ResponseForUser
}
