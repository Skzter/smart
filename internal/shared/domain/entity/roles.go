package entity

import "github.com/sashabaranov/go-openai"

const (
	// RoleUser is the role for the users messages
	RoleUser = openai.ChatMessageRoleUser
	// RoleSystem is the role for the systems messages
	RoleSystem = openai.ChatMessageRoleSystem
	// RoleAssistant is the role for the assistants messages
	RoleAssistant = openai.ChatMessageRoleAssistant
)
