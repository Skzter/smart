package entity

type RequestForLlmDTO struct {
	UserPrompt   string
	SessionID    string
	Model        string
	SystemPrompt string
}
