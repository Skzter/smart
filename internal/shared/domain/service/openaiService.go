package service

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type OpenAIService struct {
	repo   repository.OpenAI
	logger *slog.Logger
}

func NewService(logger *slog.Logger, timeout int) (*OpenAIService, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	repo, err := repository.NewOpenAiRepository(logger, timeout)
	if err != nil {
		return nil, err
	}

	return &OpenAIService{repo, logger}, nil
}

func (c *OpenAIService) Request(ctx context.Context, serviceRequest entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	switch {
	case serviceRequest.Prompt == "":
		return nil, fmt.Errorf("no prompt in request")
	case serviceRequest.SystemPrompt == "":
		return nil, fmt.Errorf("creating openAiService without system prompt")
	case serviceRequest.Model == "":
		return nil, fmt.Errorf("creating openAiService without model")
	}

	request := entity.Request{Prompt: serviceRequest.Prompt, SessionID: serviceRequest.SessionID}
	return c.repo.CreateRequest(ctx, request)
}
