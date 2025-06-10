package entity

// Response represents the response from an OpenAI API request.
type Response struct {
	Output string // The generated text output
	Id     string // Response ID for conversation tracking
	Status string // Status of the request
}
