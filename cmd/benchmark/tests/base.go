package tests

import (
	"fmt"
	"log"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BenchmarkTest interface for structered tests
type BenchmarkTest interface {
	Name() string
	Run(page playwright.Page, test *IntegrationTest) (interface{}, error)
	GetIntegrationTest() *IntegrationTest
}

// Tests global array with all tests for Benchmark
// nolint:gochecknoglobals
var (
	Tests []BenchmarkTest
)

// Register adds single test to benchmarks
func Register(test BenchmarkTest) {
	Tests = append(Tests, test)
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

// GetIntegrationTest returns integration test
func (i *IntegrationTest) GetIntegrationTest() *IntegrationTest {
	return i
}

// Name returns name of given IntegrationTest
func (i *IntegrationTest) Name() string {
	return i.testName
}

// IntegrationResult for time tracking and similarity from answers
type IntegrationResult struct {
	Duration       time.Duration
	Similarity     float64
	TestName       string
	ActualResponse string
}

// RunPrompt runs prompt and returns actual response
func RunPrompt(page playwright.Page, prompt string) (string, error) {
	promptLocator := page.Locator(`[placeholder="Prompt"]`)
	if err := promptLocator.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible}); err != nil {
		return "", fmt.Errorf("waiting for prompt input to be visible failed: %w", err)
	}

	if err := promptLocator.Fill(prompt); err != nil {
		return "", fmt.Errorf("could not fill prompt input: %w", err)
	}

	if err := page.Locator(`button:has-text("Send")`).Click(); err != nil {
		return "", fmt.Errorf("could not click Send button: %w", err)
	}

	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return "", fmt.Errorf("error waiting for response after sending prompt: %w", err)
	}

	actualResponse, err := page.Locator(`div:has(h1:has-text("Bot")) p`).Nth(1).TextContent()
	if err != nil {
		return "", fmt.Errorf("could not get bot response: %w", err)
	}

	return actualResponse, nil
}

// Login performs login for given test
func Login(page playwright.Page, b *IntegrationTest) error {
	log.Println("Performing login")
	if _, err := page.Goto(b.url); err != nil {
		return fmt.Errorf("could not goto login page: %w", err)
	}
	// Click Log in button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Log in",
	}).Click(); err != nil {
		return fmt.Errorf("could not click first login button: %w", err)
	}
	// Fill in credentials
	if err := page.Locator("#username").Fill(b.email); err != nil {
		return fmt.Errorf("could not fill email input: %w", err)
	}
	if err := page.Locator("#password").Fill(b.password); err != nil {
		return fmt.Errorf("could not fill password input: %w", err)
	}
	// Click the continue button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name:  "Continue",
		Exact: playwright.Bool(true),
	}).Click(); err != nil {
		return fmt.Errorf("could not click second login button: %w", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("error waiting for login after sending login details: %w", err)
	}
	log.Println("Login successful")
	return nil
}
