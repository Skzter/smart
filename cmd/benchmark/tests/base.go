package tests

import "github.com/playwright-community/playwright-go"

type BenchmarkTest interface {
	Name() string
	Run(page playwright.Page) (interface{}, error)
}

// nolint:gochecknoglobals
var Tests []BenchmarkTest

func Register(test BenchmarkTest) {
	Tests = append(Tests, test)
}
