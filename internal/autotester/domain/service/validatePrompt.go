package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	autotesterEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errs"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// ValidatePrompt defines an interface for prompt validation
type ValidatePrompt interface {
	ValidatePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error)
}

// validatePrompt provides functionality to validate user prompts using OpenAI.
type validatePrompt struct {
	service sharedService.OpenAI
	config  *config.Config
	logger  *slog.Logger
}

// NewValidatePromptService creates a new validatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewValidatePromptService(service sharedService.OpenAI, config *config.Config, logger *slog.Logger) (ValidatePrompt, error) {
	if err := assert.NotNil(service, config, logger); err != nil {
		return nil, err
	}
	return &validatePrompt{service, config, logger}, nil
}

// ValidatePrompt checks if the user prompt contains required information for test generation.
// It uses OpenAI service to validate the prompt against predefined validation rules.
// Returns nil if valid, ErrPromptInvalid if validation fails, or other errors on request failure.
func (s *validatePrompt) ValidatePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		return "", &errs.Error{
			Message:    fmt.Sprintf("SERVICE: ValidatePrompt(): %v", err),
			Underlying: err,
			Type:       errs.Private,
		}
	}

	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.ValidationPrompt,
	}

	resp, err := s.service.Request(ctx, req)
	if err != nil {
		return "", fmt.Errorf("SERVICE: ValdidatePrompt() => %w", err)
	}

	llmRespone := autotesterEntity.ModelAnswerText{}
	err = json.Unmarshal([]byte(resp.Text), &llmRespone)
	if err != nil {
		return "", &errs.Error{
			Message:    fmt.Sprintf("SERVICE: ValidatePrompt() - unmarshal: %s", err.Error()),
			Underlying: err,
			Type:       errs.Private,
		}
	}
	return llmRespone.Text, nil
}
