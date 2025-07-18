package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type flightSearchTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&flightSearchTest{
		IntegrationTest: &IntegrationTest{
			testName: "Flight Search Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reisesparte.  
Base-URL: https://urlaub.check24.de/reise.  
Szenario: Flugsuche München nach Barcelona im August.  
Ablauf: Auf der Startseite wählt der Nutzer 'Flug suchen', gibt im Eingabefeld für Abflugort 'München' ein, im Feld Zielort 'Barcelona', wählt Datum über Kalender-Widget (August), klickt auf Button 'Suchen'.  
Assertions: Die URL enthält '/reise/flug', die Liste der Flüge ist sichtbar und enthält mindestens einen Eintrag.  
Testdaten/Setup: Beispiel-Flugdaten aus Fixture-Datei; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Flugsuche München nach Barcelona im August", async ({ page }) => {
  await page.goto("https://urlaub.check24.de/reise");
  await auto("Wähle 'Flug suchen'", { page, test });
  await auto("Gib im Eingabefeld für Abflugort 'München' ein", { page, test });
  await auto("Gib im Feld Zielort 'Barcelona' ein", { page, test });
  await auto("Wähle Datum über Kalender-Widget im August", { page, test });
  await auto("Klicke auf Button 'Suchen'", { page, test });

  const urlContains = await auto("Enthält die URL '/reise/flug'?", { page, test });
  expect(urlContains).toBe(true);

  const flightsListVisible = await auto("Ist die Liste der Flüge sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(flightsListVisible).toBe(true);
});
    `,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *flightSearchTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
