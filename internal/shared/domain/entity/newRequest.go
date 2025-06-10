package entity

// NewRequest creates a new Request without a session ID.
func NewRequest(model string, body RequestBody) Request {
	return Request{Model: model, Body: body}
}
