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
		testName:         "Example Test",
		ExpectedResponse: "yes",
	})
}

type ExampleTest struct {
	testName         string
	ExpectedResponse string
}

func (e *ExampleTest) Name() string {
	return e.testName
}

type ExampleResult struct {
	Duration   time.Duration
	Similarity float64
}

func (e *ExampleTest) Run(page playwright.Page) (interface{}, error) {
	// Login logic
	if _, err := page.Goto("http://localhost:8081"); err != nil {
		return nil, fmt.Errorf("could not goto: %w", err)
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	}); err != nil {
		return nil, fmt.Errorf("error waiting for page to load: %w", err)
	}

	start := time.Now()

	// New steps
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

	similarity := strutil.Similarity(actualResponse, e.ExpectedResponse, metrics.NewLevenshtein())

	return &ExampleResult{
		Duration:   time.Since(start),
		Similarity: similarity,
	}, nil
}
