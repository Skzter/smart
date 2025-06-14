package entity

// UserPrompt embeds Prompt.
// Created to enforce type safety for RequestForLLM
type SystemPrompt struct {
	Prompt
}
