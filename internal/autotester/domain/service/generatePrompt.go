package service

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
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
	chatModel      model.ChatModel
	taglistService sharedService.TaglistStorage
	config         *config.Config
	logger         *slog.Logger
	validator      Validator
	tracer         trace.Tracer
}

// NewGeneratePromptService creates a new generatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewGeneratePromptService(
	chatModel model.ChatModel,
	taglistService sharedService.TaglistStorage,
	config *config.Config,
	logger *slog.Logger,
	validator Validator,
	tracer trace.Tracer,
) (GeneratePrompt, error) {
	if err := assert.NotNil(chatModel, taglistService, config, logger, validator, tracer); err != nil {
		return nil, err
	}
	return &generatePrompt{chatModel, taglistService, config, logger, validator, tracer}, nil
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
		Messages:     chat.Filter(entity.MessageTypeGeneration),
		Model:        s.config.Model,
		SystemPrompt: prompt,
	}
	chat.LastAutoPlaywrightPrompt = prompt

	if err := s.validator.ValidateRequest(ctx, req); err != nil {
		s.logger.Error("Request validation failed", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "request validation failed")
		return "", err
	}

	// Build Eino Chain
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(s.chatModel)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		s.logger.Error("Failed to compile Eino chain", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Eino chain compilation failed")
		return "", errors.ErrInternalServer
	}

	inputMessages := toEinoMessages(prompt, req.Messages)

	respMsg, err := runnable.Invoke(ctx, inputMessages)
	if err != nil {
		s.logger.Error("OpenAI request failed via Eino", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "OpenAI service request failed")
		return "", errors.ErrInternalServer
	}

	respBody := respMsg.Content
	assistantMsg := sharedEntity.NewMessage(respBody, sharedEntity.RoleAssistant)
	chat.AddMessage(assistantMsg, entity.MessageTypeGeneration)

	if err = assert.StringNotEmpty(respBody); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Empty response body")
		s.logger.Error(err.Error())
		return "", errors.ErrGeneration
	}

	span.SetStatus(codes.Ok, "")
	return respBody, nil
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
