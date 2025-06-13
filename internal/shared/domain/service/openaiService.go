package service

import (
	"context"
	"errors"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type OpenAIService struct {
	repo         repository.OpenAI
	systemPrompt string
	model        string
	logger       *slog.Logger
}

func NewService(key string, systemPrompt string, model string, logger *slog.Logger) (*OpenAIService, error) {
	if err := assert.NotNil(logger); err != nil {
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

	return &OpenAIService{repo, systemPrompt, model, logger}, nil
}

func (c *OpenAIService) Request(ctx context.Context, serviceRequest entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	if serviceRequest.Prompt == "" {
		err := errors.New("no prompt in request")
		c.logger.Error(err.Error())
		return nil, err
	}
	request := entity.Request{Prompt: serviceRequest.Prompt, LastId: serviceRequest.LastId}
	return c.repo.CreateRequest(ctx, request, c.systemPrompt, c.model)
}
