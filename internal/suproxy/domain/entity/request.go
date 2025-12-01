package entity

// Request represents a request entity with header, tags, destination, and body content.
type Request struct {
	Header      map[string]string `json:"header"`
	Tags        string            `json:"tags"`
	Destination string            `json:"destination"`
	Body        string            `json:"body"`
}
