package entity

// Prompt is an abstract prompt type that contains the common elements shared by UserPrompt and SystemPrompt.
type Prompt struct {
	Content      *Content
	LanguageCode string
}
