package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type hotelSearchSortedByRatingTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&hotelSearchSortedByRatingTest{
		IntegrationTest: &IntegrationTest{
			testName: "Hotel Search Test Sorted by Rating",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Hotelsuche auf Mallorca im August für 1 Woche, sortiert nach Bewertung.
Ablauf: Auf der Startseite wählt der Nutzer 'Nur Hotel', gibt im Eingabefeld für 'Reiseziel / Hotel' 'Mallorca' ein, wählt Anreise im August über ein Kalender-Widget, eine Reisedauer von 1 Woche und klickt auf den Button 'suchen'. Auf der Ergebnisseite wählt der Nutzer die Sortierung 'Bewertung'.
Assertions: Die URL enthält 'offerSort=relevance'. Die Liste der Hotels ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Hotelsuche auf Mallorca im August für 1 Woche, sortiert nach Bewertung", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Nur Hotel'", { page, test });
  await auto("Gib im Eingabefeld für 'Reiseziel / Hotel' 'Mallorca' ein", { page, test });
  await auto("Wähle Anreise im August über ein Kalender-Widget", { page, test });
  await auto("Wähle eine Reisedauer von 1 Woche", { page, test });
  await auto("Klicke auf den Button 'suchen'", { page, test });
  await auto("Wähle die Sortierung 'Bewertung'", { page, test });

  const url = page.url();
  expect(url).toContain("offerSort=relevance");

  const listVisible = await auto("Ist die Liste der Hotels sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *hotelSearchSortedByRatingTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
