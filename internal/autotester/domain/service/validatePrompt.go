package service

import (
	"context"
	"errors"
	"fmt"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// ValidatePrompt checks if the user prompt contains required information for test generation.
// It uses OpenAI service to validate the prompt against predefined validation rules.
// Returns nil if valid, ErrPromptInvalid if validation fails, or other errors on request failure.
func ValidatePrompt(ctx context.Context, service *service.OpenAIService, config *config.Config, userPrompt string, sessionID string) error {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        "gpt-4.1-nano-2025-04-14",
		SystemPrompt: config.Prompts.ValidationPrompt,
	}

	resp, err := service.Request(ctx, req)
	if err != nil {
		return errors.New("failed to send prompt validation request to OpenAI")
	}

	switch resp.Text {
	case "true":
		return nil
	case "false":
		return errors.New("prompt does not contain required information for test generation")
	default:
		return fmt.Errorf("unexpected validation response: %q", resp.Text)
	}
}
