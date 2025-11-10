package entity

// Request represents a request entity with header, prompt, destination, and request content.
type Request struct {
	Tags        string `json:"prompt"`
	Destination string `json:"destination"`
	Body        string `json:"request"`
}
