package service

import (
	"log/slog"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// LLMResponseParser parses raw LLM responses and maps text and code parts
// into the existing domain entities.
type LLMResponseParser interface {
	// ParseResponse parses the raw shared response and returns a populated
	// `entity.LLMResponse` containing separated text and test code.
	// The caller inspects the returned `LLMResponse` to decide which parts
	// to use.
	ParseResponse(rawResponse *sharedEntity.Response) (*entity.LLMResponse, error)
}

type llmResponseParser struct {
	logger *slog.Logger
}

// NewLLMResponseParser creates a new LLMResponseParser with the provided
// logger. It validates input parameters and returns an initialized parser
// instance or an error.
func NewLLMResponseParser(logger *slog.Logger) (LLMResponseParser, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	return &llmResponseParser{
		logger: logger,
	}, nil
}

func (l llmResponseParser) ParseResponse(rawResponse *sharedEntity.Response) (*entity.LLMResponse, error) {
	panic("unimplemented")
}
