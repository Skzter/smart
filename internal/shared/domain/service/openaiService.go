package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

type OpenAIService struct {
	context      context.Context
	api          repository.OpenAIInterface
	systemPrompt string
	model        string
	logger       *slog.Logger
}

func NewService(context context.Context, api repository.OpenAIInterface, systemPrompt string, model string, logger *slog.Logger) OpenAIService {
	return OpenAIService{context, api, systemPrompt, model, logger}
}

func (c *OpenAIService) Request(serviceRequest entity.Request) (entity.Response, error) {
	request := entity.NewRequest(serviceRequest.Prompt, serviceRequest.Id)

	return c.api.CreateRequest(context.Background(), request, c.systemPrompt, c.model)
}
