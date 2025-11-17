package entity

// Request represents a request entity with header, prompt, destination, and request content.
type Request struct {
	Header      map[string]string `json:"header"`
	Tags        string            `json:"tags"`
	Destination string            `json:"destination"`
	Body        string            `json:"body"`
}
