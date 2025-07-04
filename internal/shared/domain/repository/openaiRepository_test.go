package repository

import (
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	// "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
)

// nicht auf errorMsg prüfen, sondern nur error und im namen case beschreiben
// Test for creating new OpenAiRepository
func TestOpenaiRepository_NewOpenAiRepo(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name             string
		logger           *slog.Logger
		timeout          int
		expectedOutcome  any // hier muss mal openAi repo hin oder weg
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

func TestOpenAiRepo_ValidateRequestEntity(t *testing.T) {
	tests := []struct {
		name             string
		request          entity.Request
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "validating incorrect request entity => empty prompt",
			request: entity.Request{
				Prompt:       "",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError:    true,
			expectedErrorMsg: "request without user prompt",
		},
		{
			name: "validating incorrect request entity => empty model",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "",
				SystemPrompt: "sys prompt",
			},
			expectedError:    true,
			expectedErrorMsg: "request without model",
		},
		{
			name: "validating incorrect request entity => empty system prompt",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "",
			},
			expectedError:    true,
			expectedErrorMsg: "request without system prompt",
		},
		{
			name: "validating correct request entity",
			request: entity.Request{
				Prompt:       "user prompt",
				SessionID:    "123",
				Model:        "nano",
				SystemPrompt: "sys prompt",
			},
			expectedError:    false,
			expectedErrorMsg: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateRequestEntity(test.request)
			if test.expectedError {
				if err == nil {
					t.Errorf("expected error %q, but got nil", test.expectedErrorMsg)
				}
				if test.expectedErrorMsg != err.Error() {
					t.Errorf("wanted: %q, got: %q", test.expectedErrorMsg, err.Error())
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
// needs mock to do that, only with correct request because validation of request entity already tested
func TestOpenaiRepository_CreateRequest(t *testing.T) {
	mockRepo := mocks.NewMockOpenAI(t)
	mockRepo.EXPECT().CreateRequest()
}
*/
