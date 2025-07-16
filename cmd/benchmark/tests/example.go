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
	Register(&ExampleTest{
		testName: "Full Integration Test",
		email:    "test@autotester.com",
		password: "Autotester123",
		url:      "http://localhost:8081",
		//nolint:lll
		TestInput: `Erzeuge Playwright-Tests via Autoplaywright für Check24.
Base-URL: https://staging.check24.de.
Szenario: Nutzer-Login.
Ablauf: Der Nutzer klickt auf 'Anmelden', sieht das Eingabefeld für E-Mail mit Platzhalter 'E-Mail-Adresse' und ein Passwortfeld, gibt gültige Zugangsdaten ein (ENV-Variablen TEST_USER/TEST_PASS) und klickt auf 'Login'.
Assertions: Nach erfolgreichem Login wird die Dashboard-Seite geladen, die URL enthält '/dashboard' und der Text 'Willkommen' erscheint.
Testdaten/Setup: Stelle sicher, dass der Testuser existiert; Teardown: Logout und Session-Bereinigung.`,
		ExpectedResponse: "true",
	})
}

// ExampleTest struct with test parameters
type ExampleTest struct {
	testName         string
	email            string
	password         string
	url              string
	TestInput        string
	ExpectedResponse string
}

// Name returns name of given test
func (e *ExampleTest) Name() string {
	return e.testName
}

// ExampleResult for time tracking and similarity from answers
type ExampleResult struct {
	Duration   time.Duration
	Similarity float64
}

// Run runs playwright test => integration test so login, prompting
func (e *ExampleTest) Run(page playwright.Page) (interface{}, error) {
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

	similarity := strutil.Similarity(actualResponse, e.ExpectedResponse, metrics.NewLevenshtein())

	return &ExampleResult{
		Duration:   time.Since(start),
		Similarity: similarity,
	}, nil
}
