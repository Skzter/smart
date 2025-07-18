package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type rentalCarSearchTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&rentalCarSearchTest{
		IntegrationTest: &IntegrationTest{
			testName: "Rental Car Search Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reisesparte.
Base-URL: https://urlaub.check24.de
Szenario: Mietwagensuche in Lissabon im Oktober.
Ablauf: Auf der Startseite wählt der Nutzer 'Mietwagen', gibt im Eingabefeld für 'Abholort' 'Lissabon' ein, wählt Datum für Abholung und Rückgabe im Oktober über ein Kalender-Widget, und klickt auf den Button 'Mietwagen finden'.
Assertions: Die URL enthält '/mietwagen'. Die Liste der Mietwagen ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Mietwagensuche in Lissabon im Oktober", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Mietwagen'", { page, test });
  await auto("Gib im Eingabefeld für 'Abholort' 'Lissabon' ein", { page, test });
  await auto("Wähle Datum für Abholung und Rückgabe im Oktober über ein Kalender-Widget", { page, test });
  await auto("Klicke auf den Button 'Mietwagen finden'", { page, test });

  const urlContains = await auto("Enthält die URL '/mietwagen'?", { page, test });
  expect(urlContains).toBe(true);

  const listVisible = await auto("Ist die Liste der Mietwagen sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *rentalCarSearchTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
