package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	autotesterEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ValidatePrompt defines an interface for prompt validation
type ValidatePrompt interface {
	ValidatePrompt(ctx context.Context, userPrompt string) (bool, string, error)
}

// validatePrompt provides functionality to validate user prompts using OpenAI.
type validatePrompt struct {
	service sharedService.OpenAI
	config  *config.Config
	logger  *slog.Logger
	tracer  trace.Tracer
}

// NewValidatePromptService creates a new validatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewValidatePromptService(service sharedService.OpenAI, config *config.Config, logger *slog.Logger, tracer trace.Tracer) (ValidatePrompt, error) {
	if err := assert.NotNil(service, config, logger, tracer); err != nil {
		return nil, err
	}
	return &validatePrompt{service, config, logger, tracer}, nil
}

// ValidatePrompt checks if the user prompt contains required information for test generation.
// It uses OpenAI service to validate the prompt against predefined validation rules.
// Returns nil if valid, ErrPromptInvalid if validation fails, or other errors on request failure.
func (s *validatePrompt) ValidatePrompt(ctx context.Context, userPrompt string) (bool, string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	ctx, span := s.tracer.Start(ctx, "validatePrompt.ValidatePrompt")
	defer span.End()

	req := entity.Request{
		Messages:     []entity.Message{{Role: entity.RoleUser, Body: userPrompt}},
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.ValidationPrompt,
	}

	msg, err := s.service.Request(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "openai request failed")
		return false, "", errors.ErrValidation
	}

	llmResponse := autotesterEntity.LlmValidationResponse{}
	if err = json.Unmarshal([]byte(msg.Body), &llmResponse); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json unmarshalling failed")
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	span.SetStatus(codes.Ok, "")
	return llmResponse.Valid, llmResponse.Message, nil
}
