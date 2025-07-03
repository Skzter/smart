package repository

import (
	"log/slog"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func TestRepository(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	// test with invalid parameters
	if _, err := NewOpenAiRepository(nil, 0); err == nil {
		t.Error("called NewOpenAiRepository with invalid parmeters, but no error was thrown")
		return
	}

	// test with proper parameters
	repo, err := NewOpenAiRepository(logger, 5)
	if err != nil || repo == nil {
		t.Errorf("failed to create OpenAiRepository: %v", err)
		return
	}

	// test with missing parameters
	// point of this test is to pass invalid value as ctx parameter, but prevented by linter
	if _, err := repo.CreateRequest(t.Context(), entity.Request{}); err == nil {
		t.Error("created request with invalid parmeters, but no error was thrown")
		return
	}

	/*
		// test with proper parameters
		resp, err := repo.CreateRequest(t.Context(), entity.Request{Prompt: "What is the capital of France?", SessionID: ""}, "You are a helpfull AI assistant", "gpt-4.1-nano-2025-04-14")
		if err != nil || resp == nil {
			t.Errorf("failed to create request: %v", err)
			return
		}

		if resp.SessionID == "" {
			t.Errorf("response contained no id")
			return
		}

		if resp.Text == "" {
			t.Errorf("response contained no text")
			return
		}
	*/
}

/*
type input struct {
	timeout int
	logger  *slog.Logger
}

func TestOpenaiRepository(t *testing.T) {
	// need map with input, output of functions
	// TODO: implement
	// newOAIRepo(logger, timeout) returned nil, err, openAI struct
	// createRequest(context, request entity) returned nil, err, response entity
	testsNewOpenAiRepo := map[string]struct {
		input  input
		result any
	}{
		"failing to create new open ai repo": {
			input:  "",
			result: "",
		},
	}

	// need testing logic
		for name, test := range tests {
		  // test := test // NOTE: uncomment for Go < 1.22, see /doc/faq#closures_and_goroutines
		  t.Run(name, func(t *testing.T) {
		    t.Parallel()
		    if got, expected := reverse(test.input), test.result; got != expected { //wo reverse dann halt die zu testende Function ig
		      t.Fatalf("reverse(%q) returned %q; expected %q", test.input, got, expected)
	}
		  })
		}
}
*/
