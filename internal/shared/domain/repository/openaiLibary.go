package repository

import (
	"context"
	"errors"
	"log/slog"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type OpenAi struct {
	key string
}

func NewOpenAi(key string) *OpenAi {
	return &OpenAi{key: key}
}

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
	if err != nil || resp.Error.Message != "" {
		logger.Error("Failed to create request", "error", err)
		return entity.Response{}, err
	}

	if resp.Error.Message != "" {
		logger.Error("Request returned error", "error", err)
		return entity.Response{}, errors.New(resp.Error.Message)
	}

	return entity.Response{
		Output: resp.Output[0].Content[0].Text,
		Id:     resp.ID,
		Status: string(resp.Status),
	}, nil
}
