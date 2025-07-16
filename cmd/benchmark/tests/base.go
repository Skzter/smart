package tests

import "github.com/playwright-community/playwright-go"

// Interface for Benchmark Tests
type BenchmarkTest interface {
	Name() string
	Run(page playwright.Page) (interface{}, error)
}

// Array of all Benchmark Tests which gets accessed by Benchmark
// nolint:gochecknoglobals
var Tests []BenchmarkTest

// Function for registering single test to benchmarks
func Register(test BenchmarkTest) {
	Tests = append(Tests, test)
}
