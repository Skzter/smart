package service

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

func GeneratePrompt(ctx context.Context, service *service.OpenAIService, config *config.Config, userPrompt string, sessionID string) (string, error) {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        "gpt-4-1106-preview",
		SystemPrompt: config.Prompts.AutoPlaywrightPrompt,
	}

	resp, err := service.Request(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
