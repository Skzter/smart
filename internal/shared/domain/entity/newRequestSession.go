// Package entity provides domain entities for OpenAI API interactions.
package entity

// NewRequestSession creates a new Request with a session ID for conversations.
func NewRequestSession(model string, body RequestBody, session string) Request {
	return Request{Model: model, Body: body, Id: session}
}
