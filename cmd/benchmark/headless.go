package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/playwright-community/playwright-go"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/cmd/benchmark/tests"
)

// runHeadlessBenchmark runs the benchmark without the TUI and outputs a JSON-like report to stdout.
func runHeadlessBenchmark(parallelism int, iterations int, testsPerMinute int) {
	pw, err := playwright.Run()
	if err != nil {
		fmt.Printf("Error: could not run playwright: %v\n", err)
		os.Exit(1)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		fmt.Printf("Error: could not launch browser: %v\n", err)
		os.Exit(1)
	}
	defer browser.Close()

	var jobList []job
	for i := 0; i < iterations; i++ {
		for testIdx := range tests.Tests {
			jobList = append(jobList, job{testIndex: testIdx, iteration: i})
		}
	}

	jobs := make(chan job, parallelism)
	results := make(chan tea.Msg, len(jobList)) // buffer all results

	var wg sync.WaitGroup
	wg.Add(parallelism)

	// Worker pool
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				// Reusing runJob from TUI logic, but we need to handle the tea.Msg output
				results <- runJob(browser, j)
			}
		}()
	}

	// Dispatcher
	go func() {
		var jobDelay time.Duration
		if testsPerMinute > 0 {
			jobDelay = time.Minute / time.Duration(testsPerMinute)
		}

		for _, currentJob := range jobList {
			jobs <- currentJob
			if jobDelay > 0 {
				time.Sleep(jobDelay)
			}
		}
		close(jobs)
	}()

	wg.Wait()
	close(results)

	// Aggregate Results
	report := BenchmarkReport{
		TotalTests: len(jobList),
	}

	for msg := range results {
		switch m := msg.(type) {
		case testResultMsg:
			switch r := m.result.(type) {
			case *tests.IntegrationResult:
				report.SuccessfulTests++
				fmt.Printf("Job finished: Test: %s, Iteration: %d\n", r.TestName, m.iteration+1)
			case errMsg:
				failMsg := fmt.Sprintf("Job failed: test '%s', iteration %d, error: %v", tests.Tests[m.testIndex].Name(), m.iteration+1, r.err)
				report.Failures = append(report.Failures, failMsg)
				fmt.Println(failMsg)
			}
		}
	}

	if report.TotalTests > 0 {
		report.SuccessRate = (float64(report.SuccessfulTests) / float64(report.TotalTests)) * 100.0
	}

	fmt.Printf("\n=== Benchmark Report ===\n")
	fmt.Printf("Total: %d, Success: %d, Rate: %.2f%%\n", report.TotalTests, report.SuccessfulTests, report.SuccessRate)
	if len(report.Failures) > 0 {
		fmt.Println("Failures:")
		for _, f := range report.Failures {
			fmt.Println(f)
		}
	}
}
