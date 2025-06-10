package entity

// Request represents a complete request to the OpenAI API.
type Request struct {
	Model string      // The OpenAI model to use
	Body  RequestBody // The request body containing prompts
	Id    string      // Optional session ID for conversation continuity
}
