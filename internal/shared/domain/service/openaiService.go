package service

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	repo   repository.OpenAI
	tracer trace.Tracer
}

// NewOpenAI creates and returns a new OpenAIService instance.
func NewOpenAI(repo repository.OpenAI, tracer trace.Tracer) (OpenAI, error) {
	if err := assert.NotNil(repo); err != nil {
		return nil, err
	}
	return &openAI{repo, tracer}, nil
}

// Request sends a request to the OpenAI repository and returns the response.
func (c *openAI) Request(ctx context.Context, request entity.Request) (*entity.Message, error) {
	ctx, span := c.tracer.Start(ctx, "openAI.Request")
	defer span.End()

	if err := assert.NotNil(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "context validation failed")
		return nil, err
	}

	resp, err := c.repo.CreateRequest(ctx, request)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repository request failed")
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}
