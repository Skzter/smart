package entity

// RequestForLLM represents a request sent to the language model, including prompts and context.
type RequestForLLM struct {
	SessionId
	RequestId          string
	SystemPrompt       *SystemPrompt
	UserPrompt         *UserPrompt
	SessionContextData *SessionSummary // data to generate context for the LLM
}

func (r RequestForLLM) ToDTO() RequestForLlmDTO {
	return RequestForLlmDTO{}
}
