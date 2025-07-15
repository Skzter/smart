package handler

/*
import (
	"log/slog"
	"os"
	"testing"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
)

func TestNewAutoTesterController(t *testing.T) {
	tests := []struct {
		testName      string
		logger        *slog.Logger
		config        config.Config
		expectedError bool
	}{
		{
			testName: "valid Service",
			logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			config: config.Config{
				Model: "gpt-4",
				Port:  "8081",
				Prompts: &config.Prompts{
					ValidationPrompt:     "Bitte überprüfe die Eingabe auf Vollständigkeit.",
					AutoPlaywrightPrompt: "Erstelle automatisch ein Playwright-Skript für den folgenden Use Case.",
				},
				Timeout: 30,
				Region:  "us-central1",
				Bucket:  "my-app-bucket",
			},
			expectedError: false,
		}, {
			testName: "nil-tester",
			logger:   nil,
			config: config.Config{
				Model: "gpt-4",
				Port:  "8081",
				Prompts: &config.Prompts{
					ValidationPrompt:     "Bitte überprüfe die Eingabe auf Vollständigkeit.",
					AutoPlaywrightPrompt: "Erstelle automatisch ein Playwright-Skript für den folgenden Use Case.",
				},
				Timeout: 30,
				Region:  "us-central1",
				Bucket:  "my-app-bucket",
			},
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			controller, err := NewAutotesterController(test.logger, &test.config)

			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil for test case: %s", test.testName)
				}
				if controller != nil {
					t.Errorf("Expected nil controller, got %+v", controller)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if controller == nil {
					t.Errorf("Expected controller instance, got nil")
				}
			}
		})
	}
}

func BenchmarkRequestHandler(b *testing.B) {

}
*/
