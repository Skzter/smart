package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

const defaultValidationSystemPrompt = "Überprüfe diesen Prompt auf Richtigkeit. Antworte nur mit 'true' oder 'false'."

var ErrPromptInvalid = errors.New("prompt validation failed")

func (c *OpenAIService) ValidatePrompt(ctx context.Context, userPrompt string, sessionID string) error {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        "gpt-4.1-nano-2025-04-14",
		SystemPrompt: defaultValidationSystemPrompt,
	}

	resp, err := c.Request(ctx, req)
	if err != nil {
		return err
	}

	switch resp.Text {
	case "true":
		return nil
	case "false":
		return ErrPromptInvalid
	default:
		return fmt.Errorf("unexpected validation response: %q", resp.Text)
	}
}
