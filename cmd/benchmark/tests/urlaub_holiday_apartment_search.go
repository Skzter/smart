package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type holidayApartmentSearchTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&holidayApartmentSearchTest{
		IntegrationTest: &IntegrationTest{
			testName: "Holiday Apartment Search Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reisesparte.
Base-URL: https://urlaub.check24.de
Szenario: Ferienwohnungssuche an der Ostsee im August.
Ablauf: Auf der Startseite wählt der Nutzer 'Ferienwohnung', gibt im Eingabefeld für 'Reiseziel, Region oder Ort' 'Ostsee' ein, wählt Anreise und Abreise im August über ein Kalender-Widget, und klickt auf den Button 'Unterkünfte finden'.
Assertions: Die URL enthält '/ferienwohnung'. Die Liste der Ferienwohnungen ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";

test("Ferienwohnungssuche an der Ostsee im August", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Ferienwohnung'", { page, test });
  await auto("Gib im Eingabefeld für 'Reiseziel, Region oder Ort' 'Ostsee' ein", { page, test });
  await auto("Wähle Anreise und Abreise im August über ein Kalender-Widget", { page, test });
  await auto("Klicke auf den Button 'Unterkünfte finden'", { page, test });

  const urlContains = await auto("Enthält die URL '/ferienwohnung'?", { page, test });
  expect(urlContains).toBe(true);

  const listVisible = await auto("Ist die Liste der Ferienwohnungen sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *holidayApartmentSearchTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
