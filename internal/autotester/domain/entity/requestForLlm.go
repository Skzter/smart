package entity

// RequestForLLM represents a request sent to the language model, including prompts and context.
type RequestForLLM struct {
	SessionId
	RequestID          string
	SystemPrompt       *SystemPrompt
	UserPrompt         *UserPrompt
	SessionContextData *SessionSummary // data to generate context for the LLM
}
