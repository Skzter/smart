package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

type openAIRequest struct {
	prompt string
	lastID string
}

type OpenAIService struct {
	api          repository.RepositoryInterface
	systemPrompt string
	model        string
	logger       *slog.Logger
}

func NewRequest(prompt string) openAIRequest {
	return openAIRequest{prompt: prompt}
}

func NewRequestSession(prompt string, lastID string) openAIRequest {
	return openAIRequest{prompt: prompt, lastID: lastID}
}

func NewService(api repository.RepositoryInterface, systemPrompt string, model string, logger *slog.Logger) OpenAIService {
	return OpenAIService{api, systemPrompt, model, logger}
}

func (c *OpenAIService) Request(serviceRequest openAIRequest) (entity.Response, error) {
	c.logger.Info("Continuing conversation",
		"model", c.model,
		"prompt", serviceRequest.prompt,
		"system_prompt", c.systemPrompt,
		"previous_response_id", serviceRequest.lastID)
	request := entity.NewRequest(c.model, entity.RequestBody{UserPrompt: serviceRequest.prompt, SystemPrompt: c.systemPrompt}, serviceRequest.lastID)

	return c.api.CreateRequest(context.Background(), request)
}
