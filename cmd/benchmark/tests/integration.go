package tests

import (
	"fmt"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/playwright-community/playwright-go"
)

// nolint:gochecknoinits
func init() {
	Register(&IntegrationTest{
		testName: "Full Integration Test",
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
  await auto("Klicke auf 'Flug suchen'", { page, test });
  await auto("Gib im Eingabefeld für Abflugort 'München' ein", { page, test });
  await auto("Gib im Eingabefeld für Zielort 'Barcelona' ein", { page, test });
  await auto("Wähle im Kalender-Widget einen Termin im August", { page, test });
  await auto("Klicke auf den Button 'Suchen'", { page, test });

  const urlCorrect = await auto("Ist '/reise/flug' in der URL enthalten?", { page, test });
  expect(urlCorrect).toBe(true);

  const flightsVisible = await auto("Ist eine Liste von Flügen sichtbar mit mindestens einem Eintrag?", { page, test });
  expect(flightsVisible).toBe(true);
});
`,
	})
}

// IntegrationTest struct with test parameters
type IntegrationTest struct {
	testName         string
	email            string
	password         string
	url              string
	TestInput        string
	ExpectedResponse string
}

// Name returns name of given test
func (e *IntegrationTest) Name() string {
	return e.testName
}

// IntegrationResult for time tracking and similarity from answers
type IntegrationResult struct {
	Duration   time.Duration
	Similarity float64
}

// Run runs playwright test => integration test so login, prompting
func (e *IntegrationTest) Run(page playwright.Page) (interface{}, error) {
	start := time.Now()
	if _, err := page.Goto(e.url); err != nil {
		return nil, fmt.Errorf("could not goto: %w", err)
	}

	// Click Log in button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Log in",
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click first login button: %w", err)
	}

	// Fill in credentials
	if err := page.Locator("#username").Fill(e.email); err != nil {
		return nil, fmt.Errorf("could not fill email input: %w", err)
	}

	if err := page.Locator("#password").Fill(e.password); err != nil {
		return nil, fmt.Errorf("could not fill password input: %w", err)
	}

	// Click the continue button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name:  "Continue",
		Exact: playwright.Bool(true),
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click second login button: %w", err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("error waiting for login after sending login details: %w", err)
	}

	// Testing Prompt
	if err := page.Locator(`[placeholder="Prompt"]`).Fill(e.TestInput); err != nil {
		return nil, fmt.Errorf("could not fill prompt input: %w", err)
	}

	if err := page.Locator(`button:has-text("Send")`).Click(); err != nil {
		return nil, fmt.Errorf("could not click Send button: %w", err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("error waiting for response after sending prompt: %w", err)
	}

	// Capture and compare the response
	actualResponse, err := page.Locator(`div:has(h1:has-text("Bot")) p`).Nth(1).TextContent()
	if err != nil {
		return nil, fmt.Errorf("could not get bot response: %w", err)
	}

	similarity := strutil.Similarity(actualResponse, e.ExpectedResponse, metrics.NewJaroWinkler())

	return &IntegrationResult{
		Duration:   time.Since(start),
		Similarity: similarity,
	}, nil
}
