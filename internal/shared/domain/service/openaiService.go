package service

import (
	"context"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

type OpenAIService interface {
	Request(ctx context.Context, request entity.Request) (*entity.Response, error)
}

// openAIService handles requests to the OpenAI repository.
type openAIService struct {
	repo   repository.OpenAI
	logger *slog.Logger
}

// NewService creates and returns a new openAIService instance.
func NewService(logger *slog.Logger, timeout int) (OpenAIService, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	repo, err := repository.NewOpenAiRepository(logger, timeout)
	if err != nil {
		return nil, err
	}

	return &openAIService{repo, logger}, nil
}

// Request sends a request to the OpenAI repository and returns the response.
func (c *openAIService) Request(ctx context.Context, request entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		c.logger.Error(err.Error())
		return nil, err
	}

	return c.repo.CreateRequest(ctx, request)
}
