package service

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GeneratePrompt defines the interface for prompt generation
type GeneratePrompt interface {
	GeneratePrompt(ctx context.Context, chat *entity.Chat, request *entity.UserRequest) (string, error)
}

// generatePrompt provides functionality to generate test prompts using OpenAI.
type generatePrompt struct {
	openAIService  sharedService.OpenAI
	taglistService sharedService.TaglistStorage
	config         *config.Config
	logger         *slog.Logger
	validator      Validator
	tracer         trace.Tracer
}

// NewGeneratePromptService creates a new generatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewGeneratePromptService(
	openaiService sharedService.OpenAI,
	taglistService sharedService.TaglistStorage,
	config *config.Config,
	logger *slog.Logger,
	validator Validator,
	tracer trace.Tracer,
) (GeneratePrompt, error) {
	if err := assert.NotNil(openaiService, taglistService, config, logger, validator, tracer); err != nil {
		return nil, err
	}
	return &generatePrompt{openaiService, taglistService, config, logger, validator, tracer}, nil
}

// GeneratePrompt sends a request to OpenAI API with the provided user prompt and returns the generated response.
// It uses the AutoPlaywrightPrompt template as system prompt, filling it with tags fetched from storage.
func (s *generatePrompt) GeneratePrompt(ctx context.Context, chat *entity.Chat, request *entity.UserRequest) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrInternalServer
	}

	ctx, span := s.tracer.Start(ctx, "generatePrompt.GeneratePrompt")
	defer span.End()

	prompt := fmt.Sprintf(s.config.Prompts.AutoPlaywrightPromptT, s.formatTaglist(ctx))

	req := sharedEntity.Request{
		Messages:     chat.Filter(entity.TypeGeneration),
		Model:        s.config.Model,
		SystemPrompt: prompt,
	}
	chat.LastAutoPlaywrightPrompt = prompt

	if err := s.validator.ValidateRequest(ctx, req); err != nil {
		return "", err
	}

	resp, err := s.openAIService.Request(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "OpenAI service request failed")
		return "", err
	}
	chat.AddMessage(resp, entity.TypeGeneration|entity.TypeFrontend)

	if err = assert.StringNotEmpty(resp.Body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Empty response body")
		s.logger.Error(err.Error())
		return "", errors.ErrGeneration
	}

	span.SetStatus(codes.Ok, "")
	return resp.Body, nil
}

// formatTaglist fetches the current Taglist and formats it for the AutoPlaywrightPrompt template
func (s *generatePrompt) formatTaglist(ctx context.Context) string {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error("Context is nil, using default taglist: ", "err", err.Error())
		return sharedEntity.DefaultTagList().Format()
	}

	tagList, err := s.taglistService.GetTaglist(ctx)
	if err != nil || tagList == nil || len(tagList.Tags) == 0 {
		s.logger.Error("Failed to fetch taglist, using default: ", "err", err.Error())
		tagList = sharedEntity.DefaultTagList()
	}

	return tagList.Format()
}
