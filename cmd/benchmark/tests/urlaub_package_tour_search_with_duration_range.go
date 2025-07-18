package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type packageTourSearchWithDurationRangeTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&packageTourSearchWithDurationRangeTest{
		IntegrationTest: &IntegrationTest{
			testName: "Package Tour Search with Duration Range Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Pauschalreisesuche mit einer Reisedauer von 6-9 Tagen.
Ablauf: Auf der Startseite wählt der Nutzer 'Pauschalreisen', gibt das Reiseziel an, wählt einen Reisezeitraum, stellt die Reisedauer auf 6-9 Tage ein und klickt auf 'suchen'.
Assertions: Die URL enthält 'days=6-9'. Die Liste der Angebote ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";


test("Pauschalreisesuche mit einer Reisedauer von 6-9 Tagen", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Pauschalreisen'", { page, test });
  await auto("Gib das Reiseziel an", { page, test });
  await auto("Wähle einen Reisezeitraum", { page, test });
  await auto("Stelle die Reisedauer auf 6-9 Tage ein", { page, test });
  await auto("Klicke auf 'suchen'", { page, test });


  const urlContains = await auto("Enthält die URL 'days=6-9'?", { page, test });
  expect(urlContains).toBe(true);


  const listVisible = await auto("Ist die Liste der Angebote sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *packageTourSearchWithDurationRangeTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
