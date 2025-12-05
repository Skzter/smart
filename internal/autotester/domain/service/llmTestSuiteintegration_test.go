//go:build llmtest

package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// nolint: funlen
func PromptTestSuite(t *testing.T) {
	// Skip wenn kein API Key
	if os.Getenv("OPENAI_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_KEY not set")
	}

	ctx := context.Background()
	tracer := otel.Tracer("test")

	client := openai.NewClient(os.Getenv("OPENAI_KEY"))
	timeout := 120

	OpenAIrepo, err := repository.NewOpenAiRepository(client, timeout, tracer)
	if err != nil {
		t.Fatalf("couldn't create OpenAI repository: %v", err)
	}

	openAI, err := sharedService.NewOpenAI(OpenAIrepo, tracer)
	if err != nil {
		t.Fatalf("couldn't create OpenAI service: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("couldn't load config: %v", err)
	}
	systemPrompt := cfg.Prompts.ValidationPrompt
	model := cfg.Model

	type LLMResponse struct {
		Valid   bool   `json:"valid"`
		Message string `json:"message"`
	}

	testCases := []struct {
		name           string
		messages       []entity.Message
		expectedOutput bool
		expectError    bool
	}{
		{
			name: "Request 1 - valid with complete information",
			messages: []entity.Message{
				{
					Role: "user",
					Body: "Autoplaywright soll einen Test generieren für Check24." +
						"Base-URL: https://staging.check24.de/reise." +
						" Szenario: Flugsuche München nach Barcelona im August. " +
						"Ablauf: Auf der Startseite wählt der Nutzer 'Flug suchen', gibt Abflugort 'München' ein, Zielort 'Barcelona', wählt Datum über Kalender-Widget (August) und klickt 'Suchen'. " +
						"Assertions: URL enthält '/reise/flug', Liste der Flüge ist sichtbar und enthält mindestens einen Eintrag. " +
						"Testdaten/Setup: Beispiel-Flugdaten aus Fixture-Datei; Teardown: Browser schließen.",
				},
			},
			expectedOutput: true,
			expectError:    false,
		},
		{
			name: "Request 2 - valid with minimal information",
			messages: []entity.Message{
				{
					Role: "user",
					Body: "Autoplaywright soll einen Test erstellen für Check24." +
						" Base-URL: https://staging.check24.de/reise. " +
						"Ablauf: Nutzer sucht Flug nach Rom. " +
						"Assertions: Ergebnisse sichtbar. " +
						"Testdaten: Fixture vorhanden.",
				},
			},
			expectedOutput: true,
			expectError:    false,
		},
		{
			name: "Request 3 - invalid (empty message - caught by validation)",
			messages: []entity.Message{
				{
					Role: "user",
					Body: "",
				},
			},
			expectedOutput: false,
			expectError:    true,
		},
		{
			name: "Request 4 - invalid (missing test scenario/ablauf)",
			messages: []entity.Message{
				{
					Role: "user",
					Body: "Base-URL: https://staging.check24.de/reise. Assertions: Ergebnisse sichtbar. Testdaten: Fixture vorhanden.",
				},
			},
			expectedOutput: false,
		},
		{
			name: "Request 5 - invalid (missing ablauf details)",
			messages: []entity.Message{
				{
					Role: "user",
					Body: "Erstelle einen Autoplaywright-Test für Check24." +
						" Base-URL: https://staging.check24.de/reise." +
						" Szenario: Flugstornierung prüfen. " +
						"Assertions: Bestätigungstext erscheint, Buchung wird als storniert angezeigt. " +
						"Testdaten/Setup: Testbuchung existiert; Teardown: Buchung zurücksetzen.",
				},
			},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := entity.Request{
				Messages:     tc.messages,
				SystemPrompt: systemPrompt,
				Model:        model,
			}

			resp, err := openAI.Request(ctx, request)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else {
					t.Logf("Got expected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("OpenAI request failed: %v", err)
			}

			t.Logf("Raw response for '%s': %s", tc.name, resp.Body)

			var llmResp LLMResponse
			if err := json.Unmarshal([]byte(resp.Body), &llmResp); err != nil {
				t.Fatalf("could not unmarshal response JSON: %v\nRaw response: %s",
					err, resp.Body)
			}

			if llmResp.Valid != tc.expectedOutput {
				t.Errorf("Expected Valid=%v, got Valid=%v\nMessage: %s\nRaw: %s",
					tc.expectedOutput, llmResp.Valid, llmResp.Message, resp.Body)
			} else {
				t.Logf("Test passed: Valid=%v as expected", llmResp.Valid)
			}
		})
	}
}
