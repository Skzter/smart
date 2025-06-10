package entity

// RequestBody contains the user and system prompts for an OpenAI request.
type RequestBody struct {
	UserPrompt   string // The user's input prompt
	SystemPrompt string // The system context/instruction prompt
}
