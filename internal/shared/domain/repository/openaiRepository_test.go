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
