package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ErrNilContext is given when ctx of a request is nil
var ErrNilContext = errors.New("SERVICE: assert: request with nil context")

// OpenAI handles requests to the OpenAI repository.
type OpenAI interface {
	Request(context.Context, entity.Request) (*entity.Response, error)
}

// openAI represents a wrapper around the openai repository with a logger
type openAI struct {
	repo   repository.OpenAI
	logger *slog.Logger
}

// NewOpenAI creates and returns a new OpenAIService instance.
func NewOpenAI(logger *slog.Logger, repo repository.OpenAI) (OpenAI, error) {
	if err := assert.NotNil(logger, repo); err != nil {
		return nil, err
	}

	return &openAI{repo, logger}, nil
}

// Request sends a request to the OpenAI repository and returns the response.
func (c *openAI) Request(ctx context.Context, request entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		c.logger.Error(fmt.Sprintf("SERVICE: assertion ctx: %v", err.Error()))
		return nil, ErrNilContext
	}

	return c.repo.CreateRequest(ctx, request)
}
