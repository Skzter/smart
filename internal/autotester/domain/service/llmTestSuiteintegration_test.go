//go:build llmtest

package service

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	openai "github.com/sashabaranov/go-openai"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/repository"
	sharedService "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// nolint: funlen
func TestPromptTestSuite(t *testing.T) {
	if os.Getenv("OPENAI_KEY") == "" {
		t.Skip("Skipping integration test: OPENAI_KEY not set")
	}

	ctx := context.Background()
	tracer := otel.Tracer("test")

	apiKey := os.Getenv("OPENAI_KEY")
	client := openai.NewClient(apiKey)

	timeout := 120

	OpenAIrepo, err := repository.NewOpenAiRepository(client, timeout, tracer)
	require.NoError(t, err, "couldn't create OpenAI repository")
	require.NotNil(t, OpenAIrepo)

	openAI, err := sharedService.NewOpenAI(OpenAIrepo, tracer)
	require.NoError(t, err, "couldn't create OpenAI service")
	require.NotNil(t, openAI)

	cfg, err := config.LoadConfig()
	require.NoError(t, err, "couldn't load config")

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
						" Base-URL: https://staging/check24.de/reise." +
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
			msgs := make([]*entity.Message, len(tc.messages))
			for i := range tc.messages {
				msgs[i] = &tc.messages[i]
			}

			request := entity.Request{
				Messages:     msgs,
				SystemPrompt: systemPrompt,
				Model:        model,
			}

			resp, err := openAI.Request(ctx, request)

			if tc.expectError {
				require.Error(t, err, "expected error but got none")
				return
			}

			require.NoError(t, err, "OpenAI request failed")
			require.NotEmpty(t, resp.Body, "LLM response body empty")

			t.Logf("Raw response for '%s': %s", tc.name, resp.Body)

			var llmResp LLMResponse
			err = json.Unmarshal([]byte(resp.Body), &llmResp)
			require.NoErrorf(t, err, "could not unmarshal response JSON\nRaw response: %s", resp.Body)

			assert.Equal(t, tc.expectedOutput, llmResp.Valid,
				"Message: %s\nRaw: %s", llmResp.Message, resp.Body)
		})
	}
}
