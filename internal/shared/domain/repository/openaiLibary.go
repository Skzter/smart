package OpenAIRepository

import (
	"context"
	"errors"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
)

type OpenAi struct {
	key string
}

func NewOpenAi(key string) OpenAi {
	return OpenAi{key: key}
}

func (qa OpenAi) CreateRequest(request entity.Request, ctx context.Context) (entity.Response, error) {
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

	resp, err := client.Responses.New(ctx, openaiRequest)
	if err != nil {
		return entity.Response{}, err
	}

	return entity.Response{
		Output: resp.Output[0].Content[0].Text,
		Id:     resp.ID,
		Status: string(resp.Status),
	}, errors.New(resp.Error.Message)
}
