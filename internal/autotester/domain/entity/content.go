package entity

// Content represents a Wrapper for possible prompt elements, including text and files in addition it has a tokenCount
type Content struct {
	Text       string  // Text contains the written part of the prompt
	Files      []*File // Files contains any files associated with the prompt
	TokenCount int     // TokenCount specifies the number of tokens within the prompt
}
