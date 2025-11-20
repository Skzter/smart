package entity

import "github.com/sashabaranov/go-openai"

const (
	RoleUser      = openai.ChatMessageRoleUser
	RoleSystem    = openai.ChatMessageRoleSystem
	RoleAssistant = openai.ChatMessageRoleAssistant
)
