package repository

import (
	"context"
	"log/slog"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository/mocks"
)

// Test for creating new OpenAiRepository
func TestOpenaiRepositoryNewOpenAiRepo(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name            string
		logger          *slog.Logger
		timeout         int
		expectedOutcome any
		expectedError   bool
	}{
		{
			name:            "creating repo with nil logger, correct timeout",
			logger:          nil,
			timeout:         5,
			expectedOutcome: nil,
			expectedError:   true,
		},
		{
			name:            "creating repo with negative timeout, correct logger",
			logger:          logger,
			timeout:         -1,
			expectedOutcome: nil,
			expectedError:   true,
		},
		{
			name:          "creating repo with correct logger, correct timeout",
			logger:        logger,
			timeout:       5,
			expectedError: false,
		},
		{
			name:            "creating repo with nil logger, negative timeout",
			logger:          nil,
			timeout:         -1,
			expectedOutcome: nil,
			expectedError:   true,
		},
	}

	mockClient := mocks.NewMockOpenAIClient(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo, err := NewOpenAiRepository(test.logger, mockClient, test.timeout)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				} else if repo != test.expectedOutcome {
					t.Errorf("wanted: %q, got: %q", test.expectedOutcome, repo)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				} else if repo == nil {
					t.Errorf("wanted repo, got: %q", repo)
				}
			}
		})
	}
}

func TestOpenAiRepoValidateRequestEntity(t *testing.T) {
	tests := []struct {
		name          string
		request       entity.Request
		expectedError bool
	}{
		{
			name: "validating incorrect request entity => empty prompt",
			request: entity.Request{
				Prompt:       "",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: true,
		},
		{
			name: "validating incorrect request entity => empty model",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "",
				SystemPrompt: "sys prompt",
			},
			expectedError: true,
		},
		{
			name: "validating incorrect request entity => empty system prompt",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "",
			},
			expectedError: true,
		},
		{
			name: "validating incorrect request entity => empty sessionId",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
		{
			name: "validating correct request entity",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateRequestEntity(test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				}
			}
		})
	}
}

// Test for creating a request to openai
// nolint:funlen
func TestOpenaiReposCreateRequest(t *testing.T) {
	model := openai.GPT4Dot1Nano20250414
	logger := slog.New(slog.DiscardHandler)
	timeout := 5

	mockClient := mocks.NewMockOpenAIClient(t)

	mockSetup := []struct {
		name           string
		openaiRequest  openai.ChatCompletionRequest
		openaiResponse openai.ChatCompletionResponse
		openaiError    error
	}{
		{
			name: "correct request with correct response",
			openaiRequest: openai.ChatCompletionRequest{
				Model: model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "sys prompt",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "user prompt",
					},
				},
			},
			openaiResponse: openai.ChatCompletionResponse{
				ID: "chatcmpl-mock123",
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "This is a mocked reply from the assistant.",
						},
						FinishReason: "stop",
					},
				},
			},
			openaiError: nil,
		},
		{
			name: "correct request with message history",
			openaiRequest: openai.ChatCompletionRequest{
				Model: model,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "sys prompt",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "what is 3 + 2?",
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "3 + 2 is 5",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "who is the highest in the room?",
					},
					{
						Role:    openai.ChatMessageRoleAssistant,
						Content: "Travis Scott",
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: "user prompt",
					},
				},
			},
			openaiResponse: openai.ChatCompletionResponse{
				ID: "chatcmpl-mock456",
				Choices: []openai.ChatCompletionChoice{
					{
						Index: 0,
						Message: openai.ChatCompletionMessage{
							Role:    "assistant",
							Content: "This is a mocked reply based on full history.",
						},
						FinishReason: "stop",
					},
				},
			},
			openaiError: nil,
		},
	}

	tests := []struct {
		name          string
		request       entity.Request
		ctx           context.Context
		expectedError bool
	}{
		{
			name: "valid",
			ctx:  t.Context(),
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        model,
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
		{
			name: "valid, storing new respid",
			ctx:  t.Context(),
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "",
				Model:        model,
				SystemPrompt: "sys prompt",
			},
			expectedError: false,
		},
		{
			name:          "nil entity",
			request:       entity.Request{},
			expectedError: true,
			ctx:           t.Context(),
		},
		{
			name: "nil context",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "",
				Model:        model,
				SystemPrompt: "sys prompt",
			},
			expectedError: true,
			ctx:           nil,
		},
	}

	// set all mocks for testcases
	for _, mc := range mockSetup {
		mockClient.On("CreateChatCompletion", mock.Anything, mc.openaiRequest).Return(mc.openaiResponse, mc.openaiError)
	}

	// Run tests on mocked function
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// create Repos everytime for mock calls to be correct
			repoEmpty := openAI{
				logger:   logger,
				client:   mockClient,
				timeout:  timeout,
				messages: []entity.Message{},
			}

			repoWithMsgs := openAI{
				logger:  logger,
				client:  mockClient,
				timeout: timeout,
				messages: []entity.Message{
					{
						Actor:       "user",
						MessageBody: "what is 3 + 2?",
					},
					{
						Actor:       "assistant",
						MessageBody: "3 + 2 is 5",
					},
					{
						Actor:       "user",
						MessageBody: "who is the highest in the room?",
					},
					{
						Actor:       "assistant",
						MessageBody: "Travis Scott",
					},
				},
			}
			_, err := repoEmpty.CreateRequest(test.ctx, test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error but go nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				}
			}
			_, err = repoWithMsgs.CreateRequest(test.ctx, test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error but go nil")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				}
			}
		})
	}
}
