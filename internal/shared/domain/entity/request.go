package entity

// Request represents a complete request to the OpenAI API.
type Request struct {
	Messages     []Message
	Model        string
	SystemPrompt string
}
