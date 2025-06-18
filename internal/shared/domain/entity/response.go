package entity

// Response represents the response from an OpenAI API request.
type Response struct {
	Text      string
	SessionID string // Response ID for conversation tracking
}
