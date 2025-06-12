package entity

// Request represents a complete request to the OpenAI API.
type Request struct {
	Model string      // The OpenAI model to use
	Body  RequestBody // The request body containing prompts
	Id    string      // Optional session ID for conversation continuity
}

// NewRequestSession creates a new Request with a session ID for conversations.
func NewRequest(model string, body RequestBody, session string) Request {
	return Request{Model: model, Body: body, Id: session}
}
