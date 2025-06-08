package repository

import (
	"context"
	"fmt"
	"log/slog"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

// OpenAi represents an OpenAI API client wrapper.
type OpenAi struct {
	key string // API key for authentication
}

// NewOpenAi creates a new OpenAI client instance with the provided API key.
func NewOpenAi(key string) *OpenAi {
	return &OpenAi{key: key}
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
// and a logger for error reporting.
func (qa *OpenAi) CreateRequest(request entity.Request, ctx context.Context, logger *slog.Logger) (entity.Response, error) {
	client := openai.NewClient(
		option.WithAPIKey(qa.key),
	)

	openaiRequest := responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: param.NewOpt(request.Body.UserPrompt),
		},
		Instructions: param.NewOpt(request.Body.SystemPrompt),
		Model:        request.Model,
	}
	if request.Id != "" {
		openaiRequest.PreviousResponseID = param.NewOpt(request.Id)
	}

	resp, err := client.Responses.New(ctx, openaiRequest)
	if err != nil {
		logger.Error("OpenAI API call failed", "err", err)
		return entity.Response{}, fmt.Errorf("openai request: %w", err)
	}
	if resp.Error.Message != "" {
		logger.Error("OpenAI returned error", "msg", resp.Error.Message, "code", resp.Error.Code)
		return entity.Response{}, fmt.Errorf("openai api error: %s", resp.Error.Message)
	}

	return entity.Response{
		Output: resp.Output[0].Content[0].Text,
		Id:     resp.ID,
		Status: string(resp.Status),
	}, nil
}
