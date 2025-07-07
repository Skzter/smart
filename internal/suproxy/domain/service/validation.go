package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
)

var (
	ErrResponseNil       = errors.New("response is nil")
	ErrInvalidHTTPStatus = errors.New("invalid HTTP status")
	ErrEmptyOpenAIResult = errors.New("openai returned empty result")
)

// SetOpenAIService replaces the OpenAI service implementation – used for testing with mocks.
func (v *Validator) SetOpenAIService(mock repository.OpenAI) {
	v.openAiService = mock
}

// Validator encapsulates the logic for validating supplier offer responses
// It sends up to MaxItems individual offer prompts to an OpenAI service for consistency checks
type Validator struct {
	openAiService          repository.OpenAI
	Logger                 *slog.Logger
	SystemPromptValidation string
	Model                  string
	MaxItems               int
}

// NewValidator creates a new validator service with logger and configuration
func NewValidator(logger *slog.Logger, cfg *config.Config) *Validator {
	openAiService, err := repository.NewOpenAiRepository(logger, cfg.Timeout)
	if err != nil {
		log.Fatal(err)
	}

	return &Validator{
		openAiService:          openAiService,
		Logger:                 logger,
		SystemPromptValidation: cfg.Prompts.ValidationPrompt,
		Model:                  cfg.Model,
		MaxItems:               cfg.MaxItemsPerValidation,
	}
}

// Validate processes a supplier offer response, extracts individual offers (items), and sends up to MaxItems of them
// to an OpenAI service for validation
func (v *Validator) Validate(jsonStr string) error {
	if jsonStr == "" {
		v.Logger.Debug("Validation skipped: empty JSON string")
		return errors.New("empty json string")
	}

	var resp entity.SupplierOfferResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		v.Logger.Error("Failed to unmarshal validation JSON", "error", err)
		return err
	}

	if err := assert.NotNil(&resp); err != nil {
		v.Logger.Debug("Validation skipped: Response is nil", "error", err)
		return err
	}

	if resp.HTTPStatusCode != 200 {
		err := fmt.Errorf("%w: %d", ErrInvalidHTTPStatus, resp.HTTPStatusCode)
		v.Logger.Error("Validation failed: Invalid HTTP status", "status", resp.HTTPStatusCode, "error", err)
		return err
	}

	v.Logger.Debug("Valid HTTP response. Forwarding to OpenAI...", "status", resp.HTTPStatusCode)

	ctx := context.Background()

	for i, item := range resp.Data.Items {
		if i >= v.MaxItems {
			break
		}

		prompt := strings.TrimSpace(string(item))
		if prompt == "" {
			v.Logger.Debug("Skipping empty item", "index", i)
			continue
		}

		req := entity.Request{
			Model:        v.Model,
			Prompt:       prompt,
			SystemPrompt: v.SystemPromptValidation,
		}

		result, err := v.openAiService.CreateRequest(ctx, req)
		if err != nil {
			v.Logger.Error("OpenAI request failed", "index", i, "error", err)
			return err
		}

		if strings.TrimSpace(result.Text) == "" {
			v.Logger.Error("OpenAI returned empty result", "index", i)
			return ErrEmptyOpenAIResult
		}

		v.Logger.Debug("OpenAI response received", "index", i, "response", result.Text)
	}

	return nil
}
