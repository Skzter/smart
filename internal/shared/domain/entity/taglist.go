package entity

// TagList represents a list of tags with metadata.
type TagList struct {
	Tags []Tag `json:"tags"`
}

// Tag represents a Tag with Name and Description
type Tag struct {
	Name        string
	Description string
}
