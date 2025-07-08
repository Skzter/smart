package repository

import (
	// "context"
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

// Test for creating new OpenAiRepository
func TestOpenaiRepository_NewOpenAiRepo(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name            string
		logger          *slog.Logger
		timeout         int
		expectedOutcome OpenAI
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repo, _ := NewOpenAiRepository(test.logger, test.timeout)
			if test.expectedError {
				if repo != test.expectedOutcome {
					t.Errorf("wanted: %q, got: %q", test.expectedOutcome, repo)
				}
			}
		})
	}
}

func TestOpenAiRepo_ValidateRequestEntity(t *testing.T) {
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
			err := ValidateRequestEntity(test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error, but got nil")
				}
			} else if !test.expectedError {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				}
			}
		})
	}
}

/*
// Test for creating a request to openai
func TestOpenaiRepository_CreateRequest(t *testing.T) {
	tests := []struct {
		name             string
		ctx              context.Context
		request          entity.Request
		expectedResponse *entity.Response
		expectedError    bool
	}{
		{
			name: "test correct CreateRequest call",
			ctx:  context.TODO(),
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedResponse: &entity.Response{
				Text:      "answer from llm",
				SessionID: "123",
			},
			expectedError: false,
		},
	}

	logger := slog.New(slog.DiscardHandler)
	timeout := 5
	repo := openAI{
		logger:  logger,
		client:  mocks.NewMockOpenAIClient(t),
		timeout: timeout,
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			resp, err := repo.CreateRequest(test.ctx, test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error but go nil")
				} else if resp != test.expectedResponse {
					t.Errorf("got %q, wanted %q", resp, test.expectedResponse)
				}
			} else if !test.expectedError {
				if err != nil {
					t.Errorf("did not expect error but go this: %q", err.Error())
				}
			}
		})
	}
}
*/
