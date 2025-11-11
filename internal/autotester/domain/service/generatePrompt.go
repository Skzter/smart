package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// GeneratePrompt defines the interface for prompt generation
type GeneratePrompt interface {
	GeneratePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error)
}

// generatePrompt provides functionality to generate test prompts using OpenAI.
type generatePrompt struct {
	openAIService  sharedService.OpenAI
	taglistService sharedService.TaglistStorage
	config         *config.Config
	logger         *slog.Logger
}

// NewGeneratePromptService creates a new generatePromptService instance.
// Returns an error if any required dependencies are nil.
func NewGeneratePromptService(openaiService sharedService.OpenAI, taglistService sharedService.TaglistStorage, config *config.Config, logger *slog.Logger) (GeneratePrompt, error) {
	if err := assert.NotNil(openaiService, taglistService, config, logger); err != nil {
		return nil, err
	}
	return &generatePrompt{openaiService, taglistService, config, logger}, nil
}

// GeneratePrompt sends a request to OpenAI API with the provided user prompt and returns the generated response.
// It uses the AutoPlaywrightPrompt template as system prompt, filling it with tags fetched from storage.
func (s *generatePrompt) GeneratePrompt(ctx context.Context, userPrompt string, sessionID string) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrInternalServer
	}

	prompt, err := s.fillPrompt(ctx)
	if err != nil {
		return "", err
	}

	req := sharedEntity.Request{
		Prompt:       userPrompt,
		SessionID:    sessionID,
		Model:        s.config.Model,
		SystemPrompt: prompt,
	}

	//TODO: validate request b4 sending it | Refractor to requestValidationService
	resp, err := s.openAIService.Request(ctx, req)
	if err != nil {
		return "", err
	}

	if err = assert.StringNotEmpty(resp.Text); err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrGeneration
	}

	return resp.Text, nil
}

// fillPrompt fetches the current Taglist and completes the AutoPlaywrightPrompt template from the config
func (s *generatePrompt) fillPrompt(ctx context.Context) (string, error) {
	if err := assert.NotNil(ctx); err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrInternalServer
	}

	taglist, err := s.taglistService.GetTaglist(ctx)
	if err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrInternalServer
	}

	staglist, err := json.Marshal(taglist)
	if err != nil {
		s.logger.Error(err.Error())
		return "", errors.ErrInternalServer
	}

	prompt := fmt.Sprintf(s.config.Prompts.AutoPlaywrightPromptT, string(staglist))
	return prompt, nil
}
