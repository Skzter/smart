package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
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
	CreateRequest(context.Context, entity.Request) (*entity.Response, error)
}

// openAI represents an openAI API client wrapper and logger-system
type openAI struct {
	logger  *slog.Logger // logger for Errors and Responses
	client  openai.Client
	timeout int // timeout in seconds
}

// NewOpenAiRepository creates a new OpenAI client instance with the provided API key.
func NewOpenAiRepository(logger *slog.Logger, client openai.Client, timeout int) (OpenAI, error) {
	if err := assert.NotNil(logger, client); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timout: %d seconds", timeout)
	}

	return &openAI{
		logger:  logger,
		client:  client,
		timeout: timeout,
	}, nil
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
func (qa *openAI) CreateRequest(ctx context.Context, request entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	switch {
	case request.Prompt == "":
		return nil, fmt.Errorf("creating request without invalid Data: Prompt is empty")
	case request.SystemPrompt == "":
		return nil, fmt.Errorf("creating request without system prompt")
	case request.Model == "":
		return nil, fmt.Errorf("creating request without model")
	}

	openaiRequest := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.NewOpt(request.Prompt),
		},
		Instructions: param.NewOpt(request.SystemPrompt),
		Model:        request.Model,
	}
	if request.SessionID != "" {
		openaiRequest.PreviousResponseID = param.NewOpt(request.SessionID)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(qa.timeout))
	defer cancel()

	if ctx == nil {
		return nil, fmt.Errorf("failed to generate new context")
	}

	resp, err := qa.client.Responses.New(ctx, openaiRequest)

	if err != nil {
		err := fmt.Errorf("openai request: %w", err)
		qa.logger.Error(err.Error())
		return nil, err
	}
	if resp.Error.Message != "" {
		err := fmt.Errorf("openai api error: %s", resp.Error.Message)
		qa.logger.Error(err.Error())
		return nil, err
	}
	text := resp.OutputText()
	if text == "" {
		return nil, fmt.Errorf("openai api error: Response contains no message")
	}

	return &entity.Response{
		Text:      resp.OutputText(),
		SessionID: resp.ID,
	}, nil
}
