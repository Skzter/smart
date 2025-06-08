package entity

import "time"

type SessionID struct {
	id string
}

type Session struct {
	sessionID             SessionID
	currentSessionSummary SessionSummary
	summaryHistory        []SessionSummary
	userRequests          []*UserRequest
	requestsForLLM        []*RequestForLLM
	llmResponses          []*LLMResponse
	responsesForUser      []*ResponseForUser
}

type SessionSummary struct {
	summary   string
	createdAt time.Time
}
