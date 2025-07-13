package tests

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// nolint:gochecknoinits
func init() {
}

type LoginTest struct {
	testName string
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

	// Fill in credentials
	if err := page.Locator(`[name="email"]`).Fill("test@autotester.com"); err != nil {
		return nil, fmt.Errorf("could not fill email: %w", err)
	}
	if err := page.Locator(`[name="password"]`).Fill("Autotester123"); err != nil {
		return nil, fmt.Errorf("could not fill password: %w", err)
	}

	// Click login button
	if err := page.Locator(`button[type="submit"]`).Click(); err != nil {
		return nil, fmt.Errorf("could not click login button: %w", err)
	}

	// Wait for navigation after login, assuming it goes to a new page
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("error waiting for navigation after login: %w", err)
	}

	return &LoginResult{
		Duration: time.Since(start),
	}, nil
}
