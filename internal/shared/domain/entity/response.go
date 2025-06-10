package entity

// Response represents the response from an OpenAI API request.
type Response struct {
	Text string // The generated text output
	Id   string // Response ID for conversation tracking
}
