package service

import (
	"io"
	"log/slog"
	"testing"

	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
)

func TestNewLLMResponseParser(t *testing.T) {
	tests := []struct {
		name    string
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "nil logger returns error",
			logger:  nil,
			wantErr: true,
		},
		{
			name:    "valid logger returns parser",
			logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
			wantErr: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parser, err := NewLLMResponseParser(tc.logger)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none (parser=%v)", parser)
				}
				if parser != nil {
					t.Fatalf("expected nil parser on error, got %T", parser)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parser == nil {
				t.Fatalf("expected non-nil parser")
			}
		})
	}
}

// nolint:funlen
func TestParseResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	p, err := NewLLMResponseParser(logger)
	if err != nil {
		t.Fatalf("setup parser failed: %v", err)
	}

	tests := []struct {
		name        string
		in          sharedEntity.Response
		answerText  string
		testCode    string
		language    string
		sessionID   string
		expectPanic bool
	}{
		{
			name: "code only when text starts with import",
			in: sharedEntity.Response{
				SessionID: "s-1",
				Text: `import { test, expect } from '@playwright/test';
test('example', async ({ page }) => {
  await page.goto('https://example.com');
  await expect(page).toHaveTitle(/Example/);
});`,
			},
			answerText: "",
			testCode: `import { test, expect } from '@playwright/test';
test('example', async ({ page }) => {
  await page.goto('https://example.com');
  await expect(page).toHaveTitle(/Example/);
});`,
			language:    "ts",
			sessionID:   "s-1",
			expectPanic: false,
		},
		{
			name: "plain answer text when not starting with import",
			in: sharedEntity.Response{
				SessionID: "s-2",
				Text:      "Hier ist eine normale Antwort ohne Testcode.",
			},
			answerText:  "Hier ist eine normale Antwort ohne Testcode.",
			testCode:    "",
			language:    "ts",
			sessionID:   "s-2",
			expectPanic: false,
		},
		{
			name: "empty text currently panics (document current behavior)",
			in: sharedEntity.Response{
				SessionID: "s-3",
				Text:      "",
			},
			expectPanic: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Fatalf("expected panic, got none")
					}
				}()
			}

			got, err := p.ParseResponse(&tc.in)
			if tc.expectPanic {
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatalf("expected non-nil LLMResponse")
			}

			if got.SessionId != tc.sessionID {
				t.Errorf("SessionId = %q, want %q", got.SessionId, tc.sessionID)
			}

			if got.AnswerText == nil {
				t.Fatalf("AnswerText is nil")
			}
			if got.AnswerText.Text != tc.answerText {
				t.Errorf("AnswerText.Text = %q, want %q", got.AnswerText.Text, tc.answerText)
			}

			if got.TestCode == nil {
				t.Fatalf("TestCode is nil")
			}
			if got.TestCode.Code != tc.testCode {
				t.Errorf("TestCode.Code = %q, want %q", got.TestCode.Code, tc.testCode)
			}
			if got.TestCode.Language != tc.language {
				t.Errorf("TestCode.Language = %q, want %q", got.TestCode.Language, tc.language)
			}
		})
	}
}
