package entity

// DatabaseEntry represents a database record containing a request, response, and associated tags.
type DatabaseEntry struct {
	Request  string   `json:"request"`
	Response Response `json:"response"`
	Tags     []string `json:"tags"`
}
