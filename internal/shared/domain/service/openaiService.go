package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
)

type openAIService struct {
	context      context.Context
	api          repository.OpenAI
	systemPrompt string
	model        string
	logger       *slog.Logger
}

func NewService(ctx context.Context, repo repository.OpenAI, systemPrompt string, model string, logger *slog.Logger) (*openAIService, error) {
	// if !assert.NotNil(ctx, repo, logger) {
	//	return nil, errors.New()
	// }

	return &openAIService{ctx, repo, systemPrompt, model, logger}, nil
}

func (c *openAIService) Request(serviceRequest entity.Request) (entity.Response, error) {
	if serviceRequest.Prompt == "" {
		err := fmt.Errorf("no prompt in request")
		c.logger.Error(err.Error())
		return entity.Response{}, err
	}
	request := entity.NewRequest(serviceRequest.Prompt, serviceRequest.LastId)
	return c.api.CreateRequest(context.Background(), request, c.systemPrompt, c.model)
}
