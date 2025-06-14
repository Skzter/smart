package entity

// Request represents a complete request to the OpenAI API.
type Request struct {
	Prompt    string
	SessionID string
}
