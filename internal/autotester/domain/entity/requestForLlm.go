package entity

type RequestForLLM struct {
	SessionId
	requestID          string
	systemPrompt       *SystemPrompt
	userPrompt         *UserPrompt
	sessionContextData *SessionSummary
}
