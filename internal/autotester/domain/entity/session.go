package entity

type Session struct {
	SessionId
	currentSessionSummary SessionSummary
	summaryHistory        []*SessionSummary
	userRequests          []*UserRequest
	requestsForLLM        []*RequestForLLM
	llmResponses          []*LLMResponse
	responsesForUser      []*ResponseForUser
}
