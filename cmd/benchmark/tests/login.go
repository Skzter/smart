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
	/*
		Register(&LoginTest{
			testName:         "Login test",
		})
	*/
}

type LoginTest struct {
	testName string
}

func NewTest(name string) LoginTest {
	return LoginTest{testName: name}
}

func (lt *LoginTest) Name() string {
	return lt.testName
}

type LoginResult struct {
	Duration time.Duration
}

func (lt *LoginTest) Run(page playwright.Page) (interface{}, error) {
	start := time.Now()
	if _, err := page.Goto("http://localhost:8081"); err != nil {
		return nil, fmt.Errorf("could not goto: %w", err)
	}

	// Click Log in button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Log in",
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click first login button: %w", err)
	}

	// Fill in credentials
	if err := page.Locator("#username").Fill("test@autotester.com"); err != nil {
		return nil, fmt.Errorf("could not fill email input: %w", err)
	}
	if err := page.Locator("#password").Fill("Autotester123"); err != nil {
		return nil, fmt.Errorf("could not fill password input: %w", err)
	}

	// Click the continue button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name:  "Continue",
		Exact: playwright.Bool(true),
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click second login button: %w", err)
	}

	fmt.Println("Wir sind drinnen")
	// Wait for navigation after login, assuming it goes to a new page
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("error waiting for navigation after login: %w", err)
	}

	// auskommentieren bis unten vor benchmark
	// Testing Prompt
	if err := page.Locator(`[placeholder="Prompt"]`).Fill("input test"); err != nil {
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

	similarity := strutil.Similarity(actualResponse, "yes", metrics.NewLevenshtein())
	fmt.Println("similarity:", similarity)

	return &LoginResult{
		Duration: time.Since(start),
	}, nil
}
