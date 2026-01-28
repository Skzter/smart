package entity

// ReadTestLogStreamResponse represents the response from reading test log streams.
// It contains the actual log content and metadata about the stream state.
type ReadTestLogStreamResponse struct {
	Content []LogEvent     `json:"content"`
	Meta    map[string]any `json:"meta"`
}
