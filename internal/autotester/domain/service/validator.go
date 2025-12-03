package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	autotesterEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Validator defines the interface for validating user prompts and requests.
type Validator interface {
	ValidatePrompt(ctx context.Context, userPrompt string) (bool, string, error)
	ValidateRequest(ctx context.Context, req entity.Request) error
}

// validatePrompt provides functionality to validate outcoming requests and user prompts using OpenAI.
type validator struct {
	openAIservice sharedService.OpenAI
	config        *config.Config
	logger        *slog.Logger
	tracer        trace.Tracer
}

// NewValidatorService creates a new instance of Validator.
// Returns an error if any required dependencies are nil.
func NewValidatorService(openAIservice sharedService.OpenAI, config *config.Config, logger *slog.Logger, tracer trace.Tracer) (Validator, error) {
	if err := assert.NotNil(openAIservice, config, logger, tracer); err != nil {
		return nil, err
	}

	return &validator{openAIservice, config, logger, tracer}, nil
}

// ValidatePrompt checks if the user prompt contains required information for test generation.
// It uses OpenAI service to validate the prompt against predefined validation rules.
// Returns nil if valid, ErrPromptInvalid if validation fails, or other errors on request failure.
func (s *validator) ValidatePrompt(ctx context.Context, userPrompt string) (bool, string, error) {
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

	s.logger.Debug(fmt.Sprintf("%v", req))
	if err := s.ValidateRequest(ctx, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request validation failed")
		return false, "Invalid Request", err
	}

	resp, err := s.openAIservice.Request(ctx, req)
	if err != nil {
		s.logger.Debug(err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, "openai request failed")
		return false, "", errors.ErrValidation
	}

	llmResponse := autotesterEntity.LlmValidationResponse{}
	if err = json.Unmarshal([]byte(resp.Body), &llmResponse); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json unmarshalling failed")
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	span.SetStatus(codes.Ok, "")
	return llmResponse.Valid, llmResponse.Message, nil
}

// ValidateRequest checks if the request contains all necessary fields
// before sending it to OpenAI.
func (s *validator) ValidateRequest(ctx context.Context, req entity.Request) error {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return errors.ErrInternalServer
	}

	_, span := s.tracer.Start(ctx, "validator.ValidateRequest")
	defer span.End()
	span.SetAttributes(
		attribute.String("openai.model", req.Model),
		attribute.Int("openai.message_count", len(req.Messages)),
	)

	if req.Model == "" {
		s.logger.Error("Model empty")
		err := errors.ErrValidation
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing model")
		return err
	}
	if req.SystemPrompt == "" {
		s.logger.Error("SystemPrompt empty")
		err := errors.ErrValidation
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing system prompt")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
