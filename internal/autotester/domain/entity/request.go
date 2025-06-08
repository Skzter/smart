package entity

type UserRequest struct {
	sessionID  SessionID
	logStamp   LogStamp
	userPrompt UserPrompt
}

type RequestForLLM struct {
	requestID          string
	sessionID          SessionID
	systemPrompt       SystemPrompt
	userPrompt         UserPrompt
	sessionContextData SessionSummary
}
