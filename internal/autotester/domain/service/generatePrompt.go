package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GeneratePromptService provides functionality to generate test prompts using OpenAI.
type GeneratePromptService struct {
	service service.OpenAIService
	config  *config.Config
	logger  *slog.Logger
}

// NewGeneratePromptService creates a new GeneratePromptService instance.
// Returns an error if any required dependencies are nil.
func NewGeneratePromptService(service service.OpenAIService, config *config.Config, logger *slog.Logger) (*GeneratePromptService, error) {
	if err := assert.NotNil(service, config, logger); err != nil {
		return nil, err
	}
	return &GeneratePromptService{service, config, logger}, nil
}

// GeneratePrompt sends a request to OpenAI API with the provided user prompt and returns the generated response.
// It uses the configured AutoPlaywrightPrompt as system prompt and gpt-4-1106-preview as model.
func (s *GeneratePromptService) GeneratePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error) {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        "gpt-4-1106-preview",
		SystemPrompt: s.config.Prompts.AutoPlaywrightPrompt,
	}

	resp, err := s.service.Request(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
