package service

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"

	repository "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	suproxyEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
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
func (v *Validator) Validate(data []byte) error {
	if len(data) == 0 {
		v.Logger.Debug("Validation skipped: empty JSON string")
		return errors.New("empty json string")
	}

	var resp suproxyEntity.SupplierOfferResponse
	resp.Data = data

	if err := assert.NotNil(&resp); err != nil {
		v.Logger.Debug("Validation skipped: Response is nil", "error", err)
		return err
	}

	v.Logger.Debug("Valid HTTP response. Forwarding to OpenAI...", "status", resp.HTTPStatusCode)

	var mappedData []map[string]interface{}
	if err := json.Unmarshal(data, &mappedData); err != nil {
		v.Logger.Error("Failed to unmarshal response body", "error", err)
		return err
	}

	// ctx := context.Background()

	for _, item := range mappedData {
		str, _ := json.Marshal(item)
		v.Logger.Info("item", "item", str)
		/*
			if i >= v.MaxItems {
				break
			}

			prompt := strings.TrimSpace(string(item))
			if prompt == "" {
				v.Logger.Debug("Skipping empty item", "index", i)
				continue
			}

			req := sharedEntity.Request{
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

			var validationResult OpenAIValidationResult

			err = json.Unmarshal([]byte(result.Text), &validationResult)

			if err != nil {
				v.Logger.Error("Failed to parse OpenAI response JSON", "index", i, "error", err, "response", result.Text)
				return fmt.Errorf("invalid OpenAI response format: %w", err)
			}

			if !validationResult.Valid {
				v.Logger.Error("Validation failed", "index", i, "reason", validationResult.Reason)
				return fmt.Errorf("validation failed for item %d: reasons=%v", i, validationResult.Reason)
			}

			v.Logger.Debug("OpenAI response received", "index", i, "response", result.Text)
		*/
	}

	return nil
}
