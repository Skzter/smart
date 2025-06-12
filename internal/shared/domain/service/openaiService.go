package service

import (
	"context"
	"errors"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type openAIService struct {
	context      context.Context
	api          repository.OpenAI
	systemPrompt string
	model        string
	logger       *slog.Logger
}

func NewService(ctx context.Context, key string, systemPrompt string, model string, logger *slog.Logger) (*openAIService, error) {
	if err := assert.NotNil(ctx, logger); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	var err error
	switch {
	case key == "":
		err = errors.New("creating openAiService without key")
	case systemPrompt == "":
		err = errors.New("creating openAiService without system prompt")
	case model == "":
		err = errors.New("creating openAiService without model")
	}

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	repo, err := repository.NewOpenAiRepository(logger, key)
	if err != nil {
		return nil, err
	}

	return &openAIService{ctx, repo, systemPrompt, model, logger}, nil
}

func (c *openAIService) Request(serviceRequest entity.Request) (*entity.Response, error) {
	if serviceRequest.Prompt == "" {
		err := errors.New("no prompt in request")
		c.logger.Error(err.Error())
		return nil, err
	}
	request := entity.NewRequest(serviceRequest.Prompt, serviceRequest.LastId)
	return c.api.CreateRequest(context.Background(), request, c.systemPrompt, c.model)
}
