package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type packageTourSearchByHotelIDTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&packageTourSearchByHotelIDTest{
		IntegrationTest: &IntegrationTest{
			testName: "Package Tour Search by Hotel ID Test",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Pauschalreisesuche für das Hotel mit der ID 3542.
Ablauf: Auf der Startseite wählt der Nutzer 'Pauschalreisen', gibt in das Suchfeld die Hotel-ID '3542' ein, wählt einen Reisezeitraum und klickt auf 'suchen'.
Assertions: Die URL enthält 'hotelId=3542'. Die Liste der Angebote ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";


test("Pauschalreisesuche für das Hotel mit der ID 3542", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Pauschalreisen'", { page, test });
  await auto("Gib in das Suchfeld die Hotel-ID '3542' ein", { page, test });
  await auto("Wähle einen Reisezeitraum", { page, test });
  await auto("Klicke auf 'suchen'", { page, test });


  const urlContains = await auto("Enthält die URL 'hotelId=3542'?", { page, test });
  expect(urlContains).toBe(true);


  const listVisible = await auto("Ist die Liste der Angebote sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *packageTourSearchByHotelIDTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
