package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

const defaultValidationSystemPrompt = "Überprüfe diesen Prompt auf Richtigkeit. Antworte nur mit 'true' oder 'false'."

var ErrPromptInvalid = errors.New("prompt validation failed")

func (c *OpenAIService) ValidatePrompt(ctx context.Context, sessionID, userPrompt string) error {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        "gpt-4.1-1106-preview",
		SystemPrompt: defaultValidationSystemPrompt,
	}

	resp, err := c.Request(ctx, req)
	if err != nil {
		return err
	}

	lower := strings.ToLower(resp.Text)
	switch lower {
	case "true":
		return nil
	case "false":
		c.logger.Error("Prompt validation failed (model returned false)", slog.String("sessionID", sessionID))
		return ErrPromptInvalid
	default:
		c.logger.Error("Unexpected validation response format", slog.String("response", resp.Text))
		return fmt.Errorf("unexpected validation response: %q", resp.Text)
	}
}
