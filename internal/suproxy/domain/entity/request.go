package entity

// Request represents a request entity with header, prompt, destination, and request content.
type Request struct {
	Header      map[string]string `json:"header"`
	Prompt      string            `json:"prompt"`
	Destination string            `json:"destination"`
	Request     string            `json:"request"`
}
