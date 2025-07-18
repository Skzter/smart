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
Szenario: Hotelsuche auf Mallorca im August für 1 Woche für 2 Erwachsene, sortiert nach bester Bewertung.
Ablauf: Auf der Startseite wählt der Nutzer den Reiter 'Nur Hotel', gibt im Eingabefeld 'Reiseziel / Hotel' den Ort 'Mallorca' ein, öffnet das Kalender-Widget und wählt einen beliebigen Tag im August als Anreisedatum, wählt aus dem Dropdown für die Reisedauer '1 Woche' aus, öffnet die Reisenden-Auswahl und stellt die Anzahl der Erwachsenen auf '2' ein, und klickt auf den Button mit der Beschriftung 'suchen'. Auf der Ergebnisseite wählt der Nutzer in der Sortierungs-Dropdown-Liste die Option 'Bewertung' aus.
Assertions: Die URL der Ergebnisseite muss den Parameter 'offerSort=rating' enthalten. Die Liste der gefundenen Hotels muss sichtbar sein und mindestens ein Hotel enthalten.
Testdaten/Setup: Es sind keine spezifischen Testdaten erforderlich. Der Browser sollte nach dem Test geschlossen werden.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Hotelsuche auf Mallorca im August für 1 Woche für 2 Erwachsene, sortiert nach bester Bewertung", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle den Reiter 'Nur Hotel'", { page, test });
  await auto("Gib 'Mallorca' in das Eingabefeld 'Reiseziel / Hotel' ein", { page, test });
  await auto("Öffne das Kalender-Widget und wähle einen beliebigen Tag im August als Anreisedatum", { page, test });
  await auto("Wähle aus dem Dropdown für die Reisedauer '1 Woche' aus", { page, test });
  await auto("Öffne die Reisenden-Auswahl und stelle die Anzahl der Erwachsenen auf '2' ein", { page, test });
  await auto("Klicke auf den Button mit der Beschriftung 'suchen'", { page, test });
  await auto("Wähle in der Sortierungs-Dropdown-Liste die Option 'Bewertung' aus", { page, test });

  expect(page.url()).toContain("offerSort=rating");

  const visible = await auto("Ist die Liste der gefundenen Hotels sichtbar?", { page, test });
  expect(visible).toBe(true);

  const hotelCount = await auto("Zähle die Anzahl der gefundenen Hotels", { page, test });
  expect(hotelCount).toBeGreaterThan(0);
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
