package entity

type Session struct {
	SessionId             string `json:"conversationId"`
	CurrentSessionSummary *SessionSummary
	History               []*SessionSummary
	UserRequests          []*UserRequest
	RequestsForLLM        []*RequestForLLM
	LlmResponses          []*LLMResponse
	ResponsesForUser      []*ResponseForUser
}
