package entity

// Request represents a complete request to the OpenAI API.
type Request struct {
	Prompt string // The request body containing prompts
	Id     string // Optional session ID for conversation continuity
}

// NewRequestSession creates a new Request with a session ID for conversations.
func NewRequest(prompt string, session string) Request {
	return Request{Prompt: prompt, Id: session}
}
