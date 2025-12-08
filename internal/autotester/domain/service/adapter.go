package service

import (
	"github.com/cloudwego/eino/schema"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// toEinoMessages converts shared entity messages to Eino schema messages.
func toEinoMessages(systemPrompt string, msgs []*sharedEntity.Message) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs)+1)

	// Add System Prompt
	if systemPrompt != "" {
		out = append(out, schema.SystemMessage(systemPrompt))
	}

	// Convert History
	for _, m := range msgs {
		switch m.Role {
		case sharedEntity.RoleUser:
			out = append(out, schema.UserMessage(m.Body))
		case sharedEntity.RoleAssistant:
			out = append(out, schema.AssistantMessage(m.Body, nil))
		case sharedEntity.RoleSystem:
			out = append(out, schema.SystemMessage(m.Body))
		}
	}
	return out
}
