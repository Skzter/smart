package tests

import "github.com/playwright-community/playwright-go"

// BenchmarkTest interface for structered tests
type BenchmarkTest interface {
	Name() string
	Run(page playwright.Page) (interface{}, error)
}

// Tests global array with all tests for Benchmark
// nolint:gochecknoglobals
var Tests []BenchmarkTest

// Register adds single test to benchmarks
func Register(test BenchmarkTest) {
	Tests = append(Tests, test)
}
