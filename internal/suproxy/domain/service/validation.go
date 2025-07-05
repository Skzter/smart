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
)

type Validator struct {
	Connector              repository.OpenAI
	Logger                 *slog.Logger
	SystemPromptValidation string
	Model                  string
}

func itemsString(items []json.RawMessage) string {
	var result string
	for _, item := range items {
		result += string(item) + "\n"
	}
	return strings.TrimSpace(result)
}

// NewValidator loads the configuration and returns the validator
func NewValidator(logger *slog.Logger, cfg *config.Config) *Validator {
	connector, err := repository.NewOpenAiRepository(logger, cfg.Timeout)
	if err != nil {
		log.Fatal(err)
	}

	return &Validator{
		Connector:              connector,
		Logger:                 logger,
		SystemPromptValidation: cfg.Prompts.ValidationPrompt,
		Model:                  cfg.Model,
	}
}

func (v *Validator) Validate(resp *entity.ValidationResponse) error {
	// Assertion für nil-Check
	if err := assert.NotNil(resp); err != nil {
		v.Logger.Debug("Validation skipped: Response is nil", "error", err)
		return err
	}

	// Fehler statt Warnung bei ungültigem HTTP Status
	if resp.HTTPStatusCode != 200 {
		err := fmt.Errorf("%w: %d", ErrInvalidHTTPStatus, resp.HTTPStatusCode)
		v.Logger.Error("Validation failed: Invalid HTTP status", "status", resp.HTTPStatusCode, "error", err)
		return err
	}

	v.Logger.Debug("Valid HTTP response. Forwarding to OpenAI...", "status", resp.HTTPStatusCode)

	req := entity.Request{
		Model:        v.Model,
		Prompt:       itemsString(resp.Data.Items),
		SystemPrompt: v.SystemPromptValidation,
	}

	result, err := v.Connector.CreateRequest(context.Background(), req)
	if err != nil {
		v.Logger.Error("OpenAI request failed", "error", err)
		return err
	}

	v.Logger.Debug("OpenAI response received", "response", result.Text)
	return nil
}
