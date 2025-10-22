package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

var errOpenAIGeneration = errors.New("generation of code failed")
var errUnexpected = errors.New("encounterd unexpected error")

// GeneratePrompt defines the interface for prompt generation
type GeneratePrompt interface {
	GeneratePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error)
}

// generatePrompt provides functionality to generate test prompts using OpenAI.
type generatePrompt struct {
	service sharedService.OpenAI
	config  *config.Config
	logger  *slog.Logger
}

// NewGeneratePromptService creates a new generatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewGeneratePromptService(service sharedService.OpenAI, config *config.Config, logger *slog.Logger) (GeneratePrompt, error) {
	if err := assert.NotNil(service, config, logger); err != nil {
		return nil, err
	}
	return &generatePrompt{service, config, logger}, nil
}

// GeneratePrompt sends a request to OpenAI API with the provided user prompt and returns the generated response.
// It uses the configured AutoPlaywrightPrompt as system prompt and gpt-4-1106-preview as model.
func (s *generatePrompt) GeneratePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error) {
	req := entity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        s.config.Model,
		SystemPrompt: s.config.Prompts.AutoPlaywrightPrompt,
	}

	resp, err := s.service.Request(ctx, req)
	if err != nil {
		var repoError *repository.Error
		if errors.As(err, &repoError) {
			switch repoError.Type {
			case repository.Private:
				s.logger.Error(fmt.Sprintf("SERVICE: generate: %v", err.Error()))
				return "", errOpenAIGeneration
			case repository.Public:
				return "", repoError
			}
		} else if errors.Is(err, sharedService.ErrNilContext) {
			s.logger.Error(fmt.Sprintf("SERVICE: generate: %v", err.Error()))
			return "", errOpenAIValidation
		}
		s.logger.Error(fmt.Sprintf("SERVICE: generate: unexpected error: %s", err.Error()))
		return "", errUnexpected
	}
	return resp.Text, nil
}
