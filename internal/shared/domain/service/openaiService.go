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

func NewService(systemPrompt string, model string, logger *slog.Logger) (*OpenAIService, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	switch {
	case systemPrompt == "":
		return nil, errors.New("creating openAiService without system prompt")
	case model == "":
		return nil, errors.New("creating openAiService without model")
	}

	repo, err := repository.NewOpenAiRepository(logger)
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
		return nil, err
	}
	request := entity.Request{Prompt: serviceRequest.Prompt, SessionID: serviceRequest.SessionID}
	return c.repo.CreateRequest(ctx, request, c.systemPrompt, c.model)
}
