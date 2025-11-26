package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// OpenAI defines methods for interacting with OpenAI API.
type OpenAI interface {
	// CreateRequest sends a request to OpenAI API and returns the response.
	// It takes a Request entity containing the model and prompts,
	// a context for cancellation, and a logger for error reporting.
	CreateRequest(context.Context, entity.Request) (*entity.Message, error)
}

// OpenAIClient provides function we use on the client
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// openAI represents an openAI API client wrapper and logger-system
type openAI struct {
	client  OpenAIClient
	timeout int // timeout in seconds
}

// NewOpenAiRepository creates a new OpenAI client instance with the provided API key.
func NewOpenAiRepository(client OpenAIClient, timeout int) (OpenAI, error) {
	if err := assert.NotNil(client); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timout: %d seconds", timeout)
	}

	return &openAI{
		client:  client,
		timeout: timeout,
	}, nil
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
func (qa *openAI) CreateRequest(ctx context.Context, req entity.Request) (*entity.Message, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	if err := validateRequestEntity(req); err != nil {
		return nil, err
	}

	msg := make([]openai.ChatCompletionMessage, len(req.Messages)+1)
	msg[0] = openai.ChatCompletionMessage{
		Role:    entity.RoleSystem,
		Content: req.SystemPrompt,
	}

	for i, m := range req.Messages {
		msg[i+1] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Body,
		}
	}

	// returned ctx cannot not be nil, because it will always return a ctx
	// only if ctx would be nil from the start it would panic but already asserted it wouldnt be nil
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(qa.timeout))
	defer cancel()

	resp, err := qa.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    req.Model,
			Messages: msg,
		})

	if err != nil {
		return nil, err
	}

	// check if there are responses from api
	if len(resp.Choices) == 0 {
		return nil, errors.ErrInternalServer
	}

	// first choice of all responses
	text := resp.Choices[0].Message.Content
	if text == "" {
		return nil, errors.ErrEmptyResponse
	}

	return &entity.Message{
		Id:        uuid.NewString(),
		Role:      openai.ChatMessageRoleAssistant,
		Body:      text,
		CreatedAt: time.Now(),
	}, nil
}

func validateRequestEntity(request entity.Request) error {
	for _, req := range request.Messages {
		switch {
		case req.Body == "":
			return errors.ErrNilUserPrompt
		case req.Role == "":
			return errors.ErrNilRole
		}
	}

	switch {
	case request.SystemPrompt == "":
		return errors.ErrNilSystemPrompt
	case request.Model == "":
		return errors.ErrNilModel
	default:
		return nil
	}
}
