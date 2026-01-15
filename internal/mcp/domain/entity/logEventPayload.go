package entity

// LogEventPayload represents the payload data for a log event with Type as the category and Message as the content.
type LogEventPayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
