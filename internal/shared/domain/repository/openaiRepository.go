package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	openai "github.com/sashabaranov/go-openai"

	autoentity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
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

type OpenAIClient interface {
	// function we use on the client
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// openAI represents an openAI API client wrapper and logger-system
type openAI struct {
	logger   *slog.Logger // logger for Errors and Responses
	client   OpenAIClient
	messages []autoentity.Message
	timeout  int // timeout in seconds
}

// NewOpenAiRepository creates a new OpenAI client instance with the provided API key.
func NewOpenAiRepository(logger *slog.Logger, timeout int) (OpenAI, error) {
	if err := assert.NotNil(logger); err != nil {
		return nil, err
	}

	if timeout <= 0 {
		return nil, fmt.Errorf("invalid timout: %d seconds", timeout)
	}

	return &openAI{
		logger:   logger,
		client:   openai.NewClient(build.OpenAIKey),
		messages: []autoentity.Message{},
		timeout:  timeout,
	}, nil
}

// CreateRequest sends a request to the OpenAI API and returns the response.
// It takes a Request entity containing the model and prompts, a context for cancellation,
func (qa *openAI) CreateRequest(ctx context.Context, request entity.Request) (*entity.Response, error) {
	if err := assert.NotNil(ctx); err != nil {
		return nil, err
	}

	if err := ValidateRequestEntity(request); err != nil {
		return nil, err
	}

	// add Request from user to messages of repo
	qa.messages = append(qa.messages, autoentity.Message{
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
		err := fmt.Errorf("openai request: %w", err)
		qa.logger.Error(err.Error())
		return nil, err
	}

	if request.SessionID == "" {
		request.SessionID = resp.ID
	}

	// check if there are responses from api
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai api error: Response contains no message")
	}

	// first choice of all responses
	text := resp.Choices[0].Message.Content
	if text == "" {
		return nil, fmt.Errorf("openai api error: Response contains empty message")
	}

	// append response to message array of repo
	qa.messages = append(qa.messages, autoentity.Message{
		Actor:       openai.ChatMessageRoleAssistant,
		MessageBody: text,
	})

	return &entity.Response{
		Text:      text,
		SessionID: request.SessionID,
	}, nil
}

func ValidateRequestEntity(request entity.Request) error {
	switch {
	case request.Prompt == "":
		return fmt.Errorf("request without user prompt")
	case request.SystemPrompt == "":
		return fmt.Errorf("request without system prompt")
	case request.Model == "":
		return fmt.Errorf("request without model")
	default:
		return nil
	}
}
