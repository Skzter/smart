package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type packageTourExtendedSearchTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&packageTourExtendedSearchTest{
		IntegrationTest: &IntegrationTest{
			testName: "Package Tour Extended Search Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Erweiterte Pauschalreisesuche.
Ablauf: Auf der Startseite wählt der Nutzer 'Pauschalreisen', gibt das Reiseziel an, wählt einen Reisezeitraum, aktiviert die erweiterte Suche und klickt auf 'suchen'.
Assertions: Die URL enthält 'extendedSearch=1'. Die Liste der Angebote ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";


test("Erweiterte Pauschalreisesuche", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Pauschalreisen'", { page, test });
  await auto("Gib das Reiseziel an", { page, test });
  await auto("Wähle einen Reisezeitraum", { page, test });
  await auto("Aktiviere die erweiterte Suche", { page, test });
  await auto("Klicke auf 'suchen'", { page, test });


  const urlContains = await auto("Enthält die URL 'extendedSearch=1'?", { page, test });
  expect(urlContains).toBe(true);


  const listVisible = await auto("Ist die Liste der Angebote sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *packageTourExtendedSearchTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
	start := time.Now()
	if err := Login(page, test); err != nil {
		return nil, fmt.Errorf("could not login: %w", err)
	}

	actualResponse, err := RunPrompt(page, test.TestInput)
	if err != nil {
		return nil, fmt.Errorf("could not run prompt: %w", err)
	}

	similarity := strutil.Similarity(actualResponse, test.ExpectedResponse, metrics.NewJaroWinkler())

	return &IntegrationResult{
		Duration:       time.Since(start),
		Similarity:     similarity,
		ActualResponse: actualResponse,
	}, nil
}
