package entity

// Content represents a Wrapper for possible prompt elements, including text and files in addition it has a tokenCount
type Content struct {
	Text       string
	Files      []*File
	TokenCount int
}
