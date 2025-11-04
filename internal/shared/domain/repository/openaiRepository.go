package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	openai "github.com/sashabaranov/go-openai"

	entity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedError "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

var (
	errNilUserPrompt   = errors.New("request without user prompt")
	errNilSystemPrompt = errors.New("request without system prompt")
	errNilModel        = errors.New("request without model")
)

const (
	errEmptyResponseArray = "REPO openai error: response contains no messages to choose from"
	errEmptyResponse      = "REPO openai error: chosen response message is empty"
)

// OpenAI defines methods for interacting with OpenAI API.
type OpenAI interface {
	// CreateRequest sends a request to OpenAI API and returns the response.
	// It takes a Request entity containing the model and prompts,
	// a context for cancellation, and a logger for error reporting.
	CreateRequest(context.Context, entity.Request) (*entity.Response, error)
}

// OpenAIClient provides function we use on the client
type OpenAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// openAI represents an openAI API client wrapper and logger-system
type openAI struct {
	logger   *slog.Logger // logger for Errors and Responses
	client   OpenAIClient
	messages []entity.Message
	timeout  int // timeout in seconds
}

// NewOpenAiRepository creates a new OpenAI client instance with the provided API key.
func NewOpenAiRepository(logger *slog.Logger, client OpenAIClient, timeout int) (OpenAI, error) {
	if err := assert.NotNil(logger, client); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timout: %d seconds", timeout)
	}

	return &openAI{
		logger:   logger,
		client:   client,
		messages: []entity.Message{},
		timeout:  timeout,
	}, nil
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
func (qa *openAI) CreateRequest(ctx context.Context, request entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		qa.logger.Error(fmt.Sprintf("REPO openai: %s", err.Error()))
		return nil, sharedError.ErrInternalServer
	}

	// func validates request entity and returns custom error
	if err := validateRequestEntity(request); err != nil {
		qa.logger.Error(fmt.Sprintf("REPO entity validation: %s", err.Error()))
		return nil, sharedError.ErrInternalServer
	}

	// add Request from user to messages of repo
	qa.messages = append(qa.messages, entity.Message{
		Actor:       openai.ChatMessageRoleUser,
		MessageBody: request.Prompt,
	})

	// create history with sys prompt
	chatHistory := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: request.SystemPrompt,
		},
	}

	// add all messages to history for conversation state
	for _, message := range qa.messages {
		msg := openai.ChatCompletionMessage{
			Role:    message.Actor,
			Content: message.MessageBody,
		}
		chatHistory = append(chatHistory, msg)
	}

	// returned ctx cannot not be nil, because it will always return a ctx
	// only if ctx would be nil from the start it would panic but already asserted it wouldnt be nil
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(qa.timeout))
	defer cancel()

	resp, err := qa.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    request.Model,
			Messages: chatHistory,
		})

	if err != nil {
		qa.logger.Error(fmt.Sprintf("REPO openai request: %s", err.Error()))
		return nil, sharedError.ErrInternalServer
	}

	if request.SessionID == "" {
		request.SessionID = resp.ID
	}

	// check if there are responses from api
	if len(resp.Choices) == 0 {
		qa.logger.Error(errEmptyResponseArray)
		return nil, sharedError.ErrInternalServer
	}

	// first choice of all responses
	qa.logger.Error(fmt.Sprintf("%v", resp.Choices))
	text := resp.Choices[0].Message.Content
	if text == "" {
		qa.logger.Error(errEmptyResponse)
		return nil, sharedError.ErrInternalServer
	}

	// append response to message array of repo
	qa.messages = append(qa.messages, entity.Message{
		Actor:       openai.ChatMessageRoleAssistant,
		MessageBody: text,
	})

	return &entity.Response{
		Text:      text,
		SessionID: request.SessionID,
	}, nil
}

func validateRequestEntity(request entity.Request) error {
	switch {
	case request.Prompt == "":
		return errNilUserPrompt
	case request.SystemPrompt == "":
		return errNilSystemPrompt
	case request.Model == "":
		return errNilModel
	default:
		return nil
	}
}
