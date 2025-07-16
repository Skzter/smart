package tests

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// nolint:gochecknoinits
func init() {
	/*
		Register(&LoginTest{
			testName:         "Login test",
			email: "test@autotester.com",
			password: "Autotester123",
			url: "http://localhost:8081",
		})
	*/
}

// struct for login test
type LoginTest struct {
	testName string
	email    string
	password string
	url      string
}

// return new test with given name
func NewTest(name string) LoginTest {
	return LoginTest{testName: name}
}

// return name of given test
func (lt *LoginTest) Name() string {
	return lt.testName
}

// struct for result
type LoginResult struct {
	Duration time.Duration
}

// func for running playwright test
// runs login test
func (lt *LoginTest) Run(page playwright.Page) (interface{}, error) {
	start := time.Now()
	if _, err := page.Goto(lt.url); err != nil {
		return nil, fmt.Errorf("could not goto: %w", err)
	}

	// Click Log in button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name: "Log in",
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click first login button: %w", err)
	}

	// Fill in credentials
	if err := page.Locator("#username").Fill(lt.email); err != nil {
		return nil, fmt.Errorf("could not fill email input: %w", err)
	}
	if err := page.Locator("#password").Fill(lt.password); err != nil {
		return nil, fmt.Errorf("could not fill password input: %w", err)
	}

	// Click the continue button
	if err := page.GetByRole("button", playwright.PageGetByRoleOptions{
		Name:  "Continue",
		Exact: playwright.Bool(true),
	}).Click(); err != nil {
		return nil, fmt.Errorf("could not click second login button: %w", err)
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
