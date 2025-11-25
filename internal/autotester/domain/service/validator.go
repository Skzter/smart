package service

import (
	"context"
	"encoding/json"
	"log/slog"

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
	service sharedService.OpenAI
	config  *config.Config
	logger  *slog.Logger
}

// NewValidatorService creates a new instance of Validator.
// Returns an error if any required dependencies are nil.
func NewValidatorService(service sharedService.OpenAI, config *config.Config, logger *slog.Logger) (Validator, error) {
	if err := assert.NotNil(service, config, logger); err != nil {
		return nil, err
	}
	return &validator{service, config, logger}, nil
}

// ValidatePrompt checks if the user prompt contains required information for test generation.
// It uses OpenAI service to validate the prompt against predefined validation rules.
// Returns nil if valid, ErrPromptInvalid if validation fails, or other errors on request failure.
func (s *validator) ValidatePrompt(ctx context.Context, userPrompt string) (bool, string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}

	req := entity.Request{
		Messages:     []entity.Message{{Role: entity.RoleUser, Body: userPrompt}},
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.ValidationPrompt,
	}
	if err := s.ValidateRequest(ctx, req); err != nil {
		return false, "Invalid Request", err
	}

	if err := s.ValidateRequest(ctx, req); err != nil {
		return false, "Invalid Request", err
	}

	resp, err := s.service.Request(ctx, req)
	if err != nil {
		s.logger.Error(err.Error())
		return false, "", errors.ErrValidation
	}

	llmResponse := autotesterEntity.LlmValidationResponse{}
	if err = json.Unmarshal([]byte(resp.Body), &llmResponse); err != nil {
		s.logger.Error(err.Error())
		return false, "", errors.ErrInternalServer
	}
	return llmResponse.Valid, llmResponse.Message, nil
}

// ValidateRequest checks if the request contains all necessary fields
// before sending it to OpenAI.
func (s *validator) ValidateRequest(ctx context.Context, req entity.Request) error {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return errors.ErrInternalServer
	}

	if req.Model == "" {
		s.logger.Error("Model empty")
		return errors.ErrValidation
	}
	if req.SystemPrompt == "" {
		s.logger.Error("SystemPrompt empty")
		return errors.ErrValidation
	}

	return nil
}
