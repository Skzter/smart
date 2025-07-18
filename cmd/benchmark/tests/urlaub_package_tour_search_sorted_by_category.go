package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type packageTourSearchSortedByCategoryTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&packageTourSearchSortedByCategoryTest{
		IntegrationTest: &IntegrationTest{
			testName: "Package Tour Search Sorted by Category Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Pauschalreisesuche, sortiert nach Hotelkategorie.
Ablauf: Auf der Startseite wählt der Nutzer 'Pauschalreisen', gibt das Reiseziel an, wählt einen Reisezeitraum und klickt auf 'suchen'. Auf der Ergebnisseite wählt der Nutzer die Sortierung 'Kategorie'.
Assertions: Die URL enthält 'sorting=categoryDistribution'. Die Liste der Angebote ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";


test("Pauschalreisesuche, sortiert nach Hotelkategorie", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Pauschalreisen'", { page, test });
  await auto("Gib das Reiseziel an", { page, test });
  await auto("Wähle einen Reisezeitraum", { page, test });
  await auto("Klicke auf 'suchen'", { page, test });
  await auto("Wähle die Sortierung 'Kategorie' auf der Ergebnisseite", { page, test });


  const urlContains = await auto("Enthält die URL 'sorting=categoryDistribution'?", { page, test });
  expect(urlContains).toBe(true);


  const listVisible = await auto("Ist die Liste der Angebote sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *packageTourSearchSortedByCategoryTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
