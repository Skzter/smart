package service

import (
	"log/slog"
	"os"
	"regexp"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func TestService(t *testing.T) {
	// test with invalid parameters
	if _, err := NewService("", "", nil); err == nil {
		t.Error("service created with invalid parameters, but no error was thrown")
	}

	// test with valid parameters
	serv, err := NewService("You are a helpful ai assistant", "gpt-4.1-nano-2025-04-14", slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err != nil || serv == nil {
		t.Errorf("failed to create service: %v", err)
		return
	}

	// test with invalid parameters
	if _, err := serv.Request(t.Context(), entity.Request{}); err == nil {
		t.Error("request created with invalid parameters, but no error was thrown")
	}

	// test with valid parameters, without lastId
	resp, err := serv.Request(t.Context(), entity.Request{Prompt: "The secret number is 13"})
	if err != nil || resp == nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	if resp.Id == "" {
		t.Errorf("response contained no id")
		return
	}
	if resp.Text == "" {
		t.Errorf("response contained no text")
		return
	}

	// test follow up with context
	resp, err = serv.Request(t.Context(), entity.Request{Prompt: "What is the secret number? Send only the number in your reply", LastId: resp.Id})
	if err != nil || resp == nil {
		t.Errorf("failed to create request: %v", err)
		return
	}
	if resp.Id == "" {
		t.Errorf("response contained no id")
		return
	}
	if resp.Text == "" {
		t.Errorf("response contained no text")
		return
	}
	want := regexp.MustCompile(".*13.*")
	if !want.MatchString(resp.Text) {
		t.Errorf("conversation may not link correctly: expected number 13, got %v", resp.Text)
		return
	}
}
