package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

type hotelSearch2Weeks2ChildrenTest struct {
	*IntegrationTest
}

// nolint:gochecknoinits
func init() {
	Register(&hotelSearch2Weeks2ChildrenTest{
		IntegrationTest: &IntegrationTest{
			testName: "Hotel Search Test 2 Weeks 2 Children",
			email:    "test@autotester.com",
			password: "Autotester123",
			url:      "http://localhost:8081",
			//nolint:lll
			TestInput: `
Autoplaywright soll einen Test generieren für Check24 Reise.
Base-URL: https://urlaub.check24.de
Szenario: Hotelsuche auf Mallorca vom 01.08.2025 bis 31.10.2025 für 2 Erwachsene und 2 Kinder.
Ablauf: Auf der Startseite wählt der Nutzer 'Nur Hotel', gibt im Eingabefeld für 'Reiseziel / Hotel' 'Mallorca' ein, wählt Anreise im August über ein Kalender-Widget, eine Reisedauer von 2 Wochen, fügt 2 Kinder hinzu und klickt auf den Button 'suchen'.
Assertions: Die URL enthält 'areaId=586', 'transportType=flight', 'roomAllocation=A-A-C-C', 'departureDate=2025-08-01', 'returnDate=2025-10-31' und 'days=2w'. Die Liste der Hotels ist sichtbar und enthält mindestens einen Eintrag.
Testdaten/Setup: Keine besonderen Testdaten benötigt; Teardown: Browser schließen.
`,
			//nolint:lll
			ExpectedResponse: `
import { test, expect } from "@playwright/test";
import { auto } from "auto-playwright";


test("Hotelsuche auf Mallorca vom 01.08.2025 bis 31.10.2025 für 2 Erwachsene und 2 Kinder", async ({ page }) => {
  await page.goto("https://urlaub.check24.de");
  await auto("Wähle 'Nur Hotel'", { page, test });
  await auto("Gib im Eingabefeld für 'Reiseziel / Hotel' 'Mallorca' ein", { page, test });
  await auto("Wähle Anreise im August über ein Kalender-Widget", { page, test });
  await auto("Wähle eine Reisedauer von 2 Wochen", { page, test });
  await auto("Füge 2 Kinder hinzu", { page, test });
  await auto("Klicke auf den Button 'suchen'", { page, test });


  const urlContains = await auto("Enthält die URL 'areaId=586'?", { page, test });
  expect(urlContains).toBe(true);


  const urlContainsTransport = await auto("Enthält die URL 'transportType=flight'?", { page, test });
  expect(urlContainsTransport).toBe(true);


  const urlContainsRoomAllocation = await auto("Enthält die URL 'roomAllocation=A-A-C-C'?", { page, test });
  expect(urlContainsRoomAllocation).toBe(true);


  const urlContainsDepartureDate = await auto("Enthält die URL 'departureDate=2025-08-01'?", { page, test });
  expect(urlContainsDepartureDate).toBe(true);


  const urlContainsReturnDate = await auto("Enthält die URL 'returnDate=2025-10-31'?", { page, test });
  expect(urlContainsReturnDate).toBe(true);


  const urlContainsDays = await auto("Enthält die URL 'days=2w'?", { page, test });
  expect(urlContainsDays).toBe(true);


  const listVisible = await auto("Ist die Liste der Hotels sichtbar und enthält mindestens einen Eintrag?", { page, test });
  expect(listVisible).toBe(true);
});
`,
		},
	})
}

// Run runs playwright test => integration test so login, prompting
func (e *hotelSearch2Weeks2ChildrenTest) Run(page playwright.Page, test *IntegrationTest) (interface{}, error) {
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
