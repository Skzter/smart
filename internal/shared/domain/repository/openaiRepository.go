package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// OpenAI defines methods for interacting with OpenAI API.
type OpenAI interface {
	// CreateRequest sends a request to OpenAI API and returns the response.
	// It takes a Request entity containing the model and prompts,
	// a context for cancellation, and a logger for error reporting.
	CreateRequest(context.Context, entity.Request, string, string) (*entity.Response, error)
}

// openAI represents an openAI API client wrapper and logger-system
type openAI struct {
	key    string       // API key for authentication
	logger *slog.Logger // logger for Errors and Responses
	client openai.Client
}

// NewOpenAiRepository creates a new OpenAI client instance with the provided API key.
func NewOpenAiRepository(logger *slog.Logger, key string) (*openAI, error) {
	if err := assert.NotNil(logger); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if key == "" {
		err := errors.New("creating OpenAIRepository without key")
		logger.Error(err.Error())
		return nil, err
	}

	return &openAI{
		key:    key,
		logger: logger,
		client: openai.NewClient(
			option.WithAPIKey(key),
		),
	}, nil
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
func (qa *openAI) CreateRequest(ctx context.Context, request entity.Request, systemPrompt string, model string) (*entity.Response, error) {
	openaiRequest := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.NewOpt(request.Prompt),
		},
		Instructions: param.NewOpt(systemPrompt),
		Model:        model,
	}
	if request.LastId != "" {
		openaiRequest.PreviousResponseID = param.NewOpt(request.LastId)
	}

	resp, err := qa.client.Responses.New(ctx, openaiRequest)
	if err != nil {
		qa.logger.Error("OpenAI API call failed", "err", err)
		return nil, fmt.Errorf("openai request: %w", err)
	}
	if resp.Error.Message != "" {
		qa.logger.Error("OpenAI returned error", "msg", resp.Error.Message, "code", resp.Error.Code)
		return nil, fmt.Errorf("openai api error: %s", resp.Error.Message)
	}

	return &entity.Response{
		Text: resp.OutputText(),
		Id:   resp.ID,
	}, nil
}
