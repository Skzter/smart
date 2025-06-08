// Package entity provides domain entities for OpenAI API interactions.
package entity

// RequestBody contains the user and system prompts for an OpenAI request.
type RequestBody struct {
	UserPrompt   string // The user's input prompt
	SystemPrompt string // The system context/instruction prompt
}

// Request represents a complete request to the OpenAI API.
type Request struct {
	Model string      // The OpenAI model to use
	Body  RequestBody // The request body containing prompts
	Id    string      // Optional session ID for conversation continuity
}

// Response represents the response from an OpenAI API request.
type Response struct {
	Output string // The generated text output
	Id     string // Response ID for conversation tracking
	Status string // Status of the request
}

// NewRequest creates a new Request without a session ID.
func NewRequest(model string, body RequestBody) Request {
	return Request{Model: model, Body: body}
}

// NewRequestSession creates a new Request with a session ID for conversations.
func NewRequestSession(model string, body RequestBody, session string) Request {
	return Request{Model: model, Body: body, Id: session}
}
