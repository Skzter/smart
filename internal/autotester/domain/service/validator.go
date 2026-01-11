package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Validator defines the interface for validating user prompts and requests.
type Validator interface {
	ValidatePrompt(ctx context.Context, chat *entity.Chat, request *entity.UserRequest) (bool, string, error)
	ValidateRequest(ctx context.Context, req sharedEntity.Request) error
	ValidateChat(ctx context.Context, chat *entity.Chat) error
	ValidateMessage(ctx context.Context, msg *sharedEntity.Message) error
	ValidateGroup(ctx context.Context, group *entity.Group) error
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
func (s *validator) ValidatePrompt(ctx context.Context, chat *entity.Chat, request *entity.UserRequest) (bool, string, error) {
	if err := assert.NotNil(ctx, chat, request); err != nil {
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	ctx, span := s.tracer.Start(ctx, "validatePrompt.ValidatePrompt")
	defer span.End()

	chat.AddMessage(sharedEntity.NewMessage(request.Prompt, sharedEntity.RoleUser), entity.MessageTypeValidation)
	req := sharedEntity.Request{
		Messages:     chat.Filter(entity.MessageTypeValidation),
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.ValidationPrompt,
	}

	if err := s.ValidateRequest(ctx, req); err != nil {
		s.logger.Error("Request validation failed", "err", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "request validation failed")
		return false, "Invalid Request", errors.ErrValidation
	}

	resp, err := s.openAIservice.Request(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "openai request failed")
		return false, "", errors.ErrValidation
	}

	llmResponse := entity.LlmValidationResponse{}
	if err = json.Unmarshal([]byte(resp.Body), &llmResponse); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "json unmarshalling failed")
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	if llmResponse.Valid {
		chat.Messages[len(chat.Messages)-1].Type = entity.MessageTypeAny
	}

	chat.AddMessage(resp, entity.MessageTypeValidation)

	span.SetStatus(codes.Ok, "")
	return llmResponse.Valid, llmResponse.Message, nil
}

// ValidateChat validates a Chat entity.
// Returns an error if any required field is empty or invalid.
func (s *validator) ValidateChat(ctx context.Context, chat *entity.Chat) error {
	if err := assert.NotNil(ctx, chat); err != nil {
		return err
	}

	_, span := s.tracer.Start(ctx, "validator.ValidateChat")
	defer span.End()

	if err := assert.StringsNotEmpty(
		chat.Id,
		chat.Author,
		chat.LastModifiedBy,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing fields")
		return fmt.Errorf("missing fields %w", err)
	}

	if err := assert.StringsNotEmpty(
		chat.Groups...,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid groups")
		return fmt.Errorf("invalid groups %w", err)
	}

	if chat.UpdatedAt.IsZero() {
		err := fmt.Errorf("updatedAt is zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, "updatedAt zero")
		return err
	}
	if chat.CreatedAt.IsZero() {
		err := fmt.Errorf("createdAt is Zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, "createdAt zero")
		return err
	}

	if err := assert.ArrayLengthGreaterThan(chat.Messages, 0); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing messages")
		return err
	}

	for _, msg := range chat.Messages {
		if err := s.ValidateMessage(ctx, &msg.Message); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "message validation failed")
			return fmt.Errorf("invalid message: %w", err)
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ValidateRequest checks if the request contains all necessary fields
// before sending it to OpenAI.
func (s *validator) ValidateRequest(ctx context.Context, req sharedEntity.Request) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}

	_, span := s.tracer.Start(ctx, "validator.ValidateRequest")
	defer span.End()
	span.SetAttributes(
		attribute.String("openai.model", req.Model),
		attribute.Int("openai.message_count", len(req.Messages)),
	)

	if req.Model == "" {
		err := fmt.Errorf("model is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing model")
		return err
	}
	if req.SystemPrompt == "" {
		err := fmt.Errorf("systemPrompt is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing system prompt")
		return err
	}

	if len(req.Messages) == 0 {
		err := fmt.Errorf("messages is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing messages")
		return err
	}

	for _, msg := range req.Messages {
		if err := s.ValidateMessage(ctx, msg); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "message validation failed")
			return fmt.Errorf("invalid message: %w", err)
		}
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ValidateMessage validates a Message entity.
// Returns an error if any required field is empty or invalid.
func (s *validator) ValidateMessage(ctx context.Context, msg *sharedEntity.Message) error {
	if err := assert.NotNil(ctx, msg); err != nil {
		return err
	}

	_, span := s.tracer.Start(ctx, "validator.ValidateMessage")
	defer span.End()

	if msg.Body == "" {
		err := fmt.Errorf("body is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing body")
		return err
	}

	if msg.Role == "" {
		err := fmt.Errorf("role is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing role")
		return err
	}

	if err := uuid.Validate(msg.Id); err != nil {
		err := fmt.Errorf("invalid Id: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid id")
		return err
	}

	if msg.CreatedAt.IsZero() {
		err := fmt.Errorf("createdAt is Zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, "createdAt zero")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// ValidateGroup validates a Group entity.
// Returns an error if any required field is empty or invalid.
func (s *validator) ValidateGroup(ctx context.Context, group *entity.Group) error {
	if err := assert.NotNil(ctx, group); err != nil {
		return err
	}

	_, span := s.tracer.Start(ctx, "validator.ValidateGroup")
	defer span.End()

	if err := assert.StringsNotEmpty(
		group.Id,
		group.Name,
		group.CreatedBy,
	); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing fields")
		return fmt.Errorf("missing fields %w", err)
	}

	if group.CreatedAt.IsZero() {
		err := fmt.Errorf("createdAt is Zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, "createdAt zero")
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}
