package entity

// SystemPrompt represents a system prompt for the autotester.
// UserPrompt embeds Prompt.
// Created to enforce type safety for RequestForLLM
type SystemPrompt struct {
	Prompt
}
