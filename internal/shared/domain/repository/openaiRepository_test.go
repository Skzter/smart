package repository

import (
	"context"
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func TestOpenaiRepository_NewOpenAiRepo(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name             string
		logger           *slog.Logger
		timeout          int
		expectedOutcome  any // hier muss mal openAi repo hin
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:             "creating repo with nil logger, correct timeout",
			logger:           nil,
			timeout:          5,
			expectedOutcome:  nil,
			expectedError:    true,
			expectedErrorMsg: "assert failed: given value at index 0 is nil",
		},
		{
			name:             "creating repo with negative timeout, correct logger",
			logger:           logger,
			timeout:          -1,
			expectedOutcome:  nil,
			expectedError:    true,
			expectedErrorMsg: "invalid timout: -1 seconds",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.expectedError {
				repo, err := NewOpenAiRepository(test.logger, test.timeout)
				if repo != test.expectedOutcome && err.Error() != test.expectedErrorMsg {
					t.Errorf("wanted: %q, got: %q", test.expectedOutcome, repo)
				}
			}
		})
	}
}

func TestOpenaiRepository_CreateRequest(t *testing.T) {
	// helper logger and repo
	logger := slog.New(slog.DiscardHandler)
	repo, _ := NewOpenAiRepository(logger, 5)

	tests := []struct {
		name             string
		ctx              context.Context
		request          entity.Request
		expectedOutcome  *entity.Response
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "create request with nil context, correct request entity",
			ctx:  nil,
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedOutcome:  nil,
			expectedError:    true,
			expectedErrorMsg: "assert failed: given value at index 0 is nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.expectedError {
				resp, err := repo.CreateRequest(test.ctx, test.request)
				if resp != test.expectedOutcome && err.Error() != test.expectedErrorMsg {
					t.Errorf("wanted: %q, got: %q", test.expectedOutcome, resp)
				}
			}
		})
	}
}
