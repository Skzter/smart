package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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

// OpenAIValidationResult models the expected JSON structure of the OpenAI validation response
type OpenAIValidationResult struct {
	Valid  bool     `json:"valid"`
	Reason []string `json:"reason"`
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
func (v *Validator) Validate(ctx context.Context, resp *entity.SupplierOfferResponse) error {
	if err := assert.NotNil(ctx, resp); err != nil {
		return err
	}

	if len(resp.Data) == 0 {
		v.Logger.Debug("Validation skipped: empty JSON string")
		return errors.New("empty json string")
	}

	if resp.HTTPStatusCode != http.StatusOK {
		return errors.New("validation failed: given response is not 200")
	}

	v.Logger.Debug("Valid HTTP response. Forwarding to OpenAI...", "status", resp.HTTPStatusCode)

	var mappedData []map[string]any
	if err := json.Unmarshal(resp.Data, &mappedData); err != nil {
		return err
	}

	for i, offer := range mappedData {
		mOffer, _ := json.Marshal(offer)
		item := string(mOffer)

		if i >= v.MaxItems {
			break
		}

		item = strings.TrimSpace(item)
		if item == "" {
			v.Logger.Debug("Skipping empty item", "index", i)
			continue
		}

		req := sharedEntity.Request{
			Model:        v.Model,
			Prompt:       item,
			SystemPrompt: v.SystemPromptValidation,
		}

		result, err := v.openAiService.CreateRequest(ctx, req)
		if err != nil {
			return err
		}

		if strings.TrimSpace(result.Text) == "" {
			return fmt.Errorf("%w index: %d", ErrEmptyOpenAIResult, i)
		}

		var validationResult OpenAIValidationResult

		err = json.Unmarshal([]byte(result.Text), &validationResult)

		if err != nil {
			return fmt.Errorf("invalid OpenAI response format: %w index: %d", err, i)
		}

		if !validationResult.Valid {
			return fmt.Errorf("validation failed for item %d: reasons=%v", i, validationResult.Reason)
		}

		v.Logger.Debug("OpenAI response received", "index", i, "response", result.Text)
	}

	return nil
}
