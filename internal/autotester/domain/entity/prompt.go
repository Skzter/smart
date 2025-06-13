package entity

// Prompt is an abstract prompt type that contains the common elements shared by User and SystemPrompt.
type Prompt struct {
	Content      *Content // content holds the actual prompt
	LanguageCode string   // languageCode specifies the language of the prompt
	LogStamp     LogStamp // logStamp tracks the logging data
}
