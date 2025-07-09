package entity

// UserPrompt embeds Prompt.
// Created to enforce type safety for RequestForLLM
type UserPrompt struct {
	Prompt
	SessionId string `json:"conversationId"`
}
