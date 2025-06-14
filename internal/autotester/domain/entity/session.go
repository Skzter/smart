package entity

type Session struct {
	SessionId             // unique identifier of the session, embedded struct
	CurrentSessionSummary *SessionSummary
	History               []*SessionSummary
	UserRequests          []*UserRequest
	RequestsForLLM        []*RequestForLLM
	LlmResponses          []*LLMResponse
	ResponsesForUser      []*ResponseForUser
}
