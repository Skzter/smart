package entity

// ValidateMessage represents a feedback message from the Autotester API
// indicating whether a prompt is valid or requires more information.
type ValidateMessage struct {
	Body string `json:"body"`
}
