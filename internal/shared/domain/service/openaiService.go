package service

import (
	"context"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// OpenAI handles requests to the OpenAI repository.
type OpenAI interface {
	Request(context.Context, entity.Request) (*entity.Message, error)
}

// openAI represents a wrapper around the openai repository with a logger
type openAI struct {
	repo repository.OpenAI
}

// NewOpenAI creates and returns a new OpenAIService instance.
func NewOpenAI(repo repository.OpenAI) (OpenAI, error) {
	if err := assert.NotNil(repo); err != nil {
		return nil, err
	}
	return &openAI{repo}, nil
}

// Request sends a request to the OpenAI repository and returns the response.
func (c *openAI) Request(ctx context.Context, request entity.Request) (*entity.Message, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	return c.repo.CreateRequest(ctx, request)
}
