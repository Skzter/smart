package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Error is custom error type for validation service
type Error struct {
	underlying error
	message    string
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Unwrap() error {
	return e.underlying
}

// Errors which the llm validation can return
var (
	ErrNotEnoughInformation = errors.New("prompt does not contain required information for test generation")
	ErrUnexpectedResponse   = errors.New("unexpected validation response, please try again")
)

// ValidatePrompt defines an interface for prompt validation
type ValidatePrompt interface {
	ValidatePrompt(ctx context.Context, userPrompt string, sessionID string) error
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
func (s *validatePrompt) ValidatePrompt(ctx context.Context, userPrompt string, sessionID string) error {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.ValidationPrompt,
	}

	resp, err := s.service.Request(ctx, req)
	if err != nil {
		return err
	}

	switch resp.Text {
	case "true":
		return nil
	case "false":
		return &Error{
			underlying: ErrNotEnoughInformation,
			message:    fmt.Sprintf("SERVICE: ValidatePrompt() => %v", ErrNotEnoughInformation.Error()),
		}
	default:
		return &Error{
			underlying: ErrUnexpectedResponse,
			message:    fmt.Sprintf("SERVICE: ValidatePrompt(): no binary decision in validation => %s", resp.Text),
		}
	}
}
