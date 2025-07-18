package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type packageTourSearchTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&packageTourSearchTest{
		IntegrationTest: &IntegrationTest{
			testName: "Package Tour Search Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reisesparte.
Base-URL: https://urlaub.check24.de
Szenario: Pauschalreisesuche nach Kreta im September.
Ablauf: Auf der Startseite wählt der Nutzer 'Pauschalreisen', gibt im Eingabefeld für 'Reiseziel oder Hotel' 'Kreta' ein, wählt Anreise im September über ein Kalender-Widget, eine Reisedauer von 1 Woche und 2 Erwachsene, und klickt auf den Button 'suchen'.
Assertions: Die URL enthält '/pauschalreisen-ergebnis'. Die Liste der Angebote ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Pauschalreisesuche nach Kreta im September", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Pauschalreisen'", { page, test });
  await auto("Gib im Eingabefeld für 'Reiseziel oder Hotel' 'Kreta' ein", { page, test });
  await auto("Wähle Anreise im September über ein Kalender-Widget", { page, test });
  await auto("Wähle eine Reisedauer von 1 Woche", { page, test });
  await auto("Wähle 2 Erwachsene", { page, test });
  await auto("Klicke auf den Button 'suchen'", { page, test });

  const urlContains = await auto("Enthält die URL '/pauschalreisen-ergebnis'?", { page, test });
  expect(urlContains).toBe(true);

  const listVisible = await auto("Ist die Liste der Angebote sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *packageTourSearchTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
