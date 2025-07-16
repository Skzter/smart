package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	service "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/suproxy/domain/entity"
)

// Common validation errors returned by the Validator.
var (
	ErrResponseNil       = errors.New("response is nil")
	ErrInvalidHTTPStatus = errors.New("invalid HTTP status")
	ErrEmptyOpenAIResult = errors.New("openai returned empty result")
)

// OpenAIValidationResult models the expected JSON structure of the OpenAI validation response
type OpenAIValidationResult struct {
	Valid  bool     `json:"valid"`
	Reason []string `json:"reason"`
}

// Validator defines an interface for validating a SupplierResponse.
// Implementations should provide specific validation logic.
type Validator interface {
	Validate(ctx context.Context, offers *entity.SupplierResponse) error
}

// Validator encapsulates the logic for validating supplier offer responses
// It sends up to MaxItems individual offer prompts to an OpenAI service for consistency checks
type validator struct {
	openAiService service.OpenAIService
	Logger        *slog.Logger
	cfg           *config.Config
}

// NewValidator creates a new validator service with logger and configuration
func NewValidator(logger *slog.Logger, cfg *config.Config, service service.OpenAIService) (Validator, error) {
	if err := assert.NotNil(logger, cfg, service); err != nil {
		return nil, err
	}

	return &validator{
		openAiService: service,
		Logger:        logger,
		cfg:           cfg,
	}, nil
}

// Validate processes a supplier offer response, extracts individual offers (items), and sends up to MaxItems of them
// to an OpenAI service for validation
func (v validator) Validate(ctx context.Context, offers *entity.SupplierResponse) error {
	if err := assert.NotNil(ctx, offers); err != nil {
		return err
	}

	if offers.HTTPStatusCode != http.StatusOK {
		return errors.New("validation failed: given response is not 200")
	}

	if len(offers.Data.Items) == 0 {
		return errors.New("validation failed: no offers in provided list")
	}

	v.Logger.Debug("Valid offerlist. Beginning LMM validation")

	for i, offer := range offers.Data.Items {
		if i >= v.cfg.MaxItemsPerValidation {
			break
		}

		item := string(offer)

		item = strings.TrimSpace(item)
		if item == "" {
			v.Logger.Debug("Skipping empty item", "index", i)
			continue
		}

		req := shared.Request{
			Model:        v.cfg.Model,
			Prompt:       item,
			SystemPrompt: v.cfg.Prompts.ValidationPrompt,
		}

		result, err := v.openAiService.Request(ctx, req)
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
