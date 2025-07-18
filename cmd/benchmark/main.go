package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/playwright-community/playwright-go"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/cmd/benchmark/tests"
)

const (
	// for whole integration 15
	// login 90
	iterationsPerTest  = 100
	defaultParallelism = 5
)

// nolint:gochecknoglobals
var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

type job struct {
	testIndex int
	iteration int
}

type testResultMsg struct {
	testIndex int
	iteration int
	result    tea.Msg
}

type errMsg struct{ err error }

type startBenchmarkMsg struct{}

type model struct {
	pw            *playwright.Playwright
	browser       playwright.Browser
	running       bool
	spinner       spinner.Model
	progress      progress.Model
	results       []*tests.IntegrationResult
	errorMsg      string
	debug         bool
	currentAction string
	parallelism   int
	jobs          chan job
	totalJobs     int
	completedJobs int
	iterations    int
}

func initialModel(debug bool, parallelism int, iterations int) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner:     s,
		progress:    progress.New(progress.WithDefaultGradient()),
		debug:       debug,
		parallelism: parallelism,
		iterations:  iterations,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.EnterAltScreen,
		func() tea.Msg { return startBenchmarkMsg{} },
	)
}

func runJobCmd(browser playwright.Browser, j job) tea.Cmd {
	return func() tea.Msg {
		test := tests.Tests[j.testIndex]
		log.Printf("Starting job: test '%s', iteration %d", test.Name(), j.iteration+1)

		context, err := browser.NewContext(playwright.BrowserNewContextOptions{
			BaseURL: playwright.String("http://localhost:8081"),
		})
		if err != nil {
			return testResultMsg{
				testIndex: j.testIndex,
				iteration: j.iteration,
				result:    errMsg{fmt.Errorf("could not create browser context: %w", err)},
			}
		}
		defer func() {
			if err := context.Close(); err != nil {
				log.Printf("could not close browser context: %v", err)
			}
		}()

		page, err := context.NewPage()
		if err != nil {
			return testResultMsg{
				testIndex: j.testIndex,
				iteration: j.iteration,
				result:    errMsg{fmt.Errorf("could not create page: %w", err)},
			}
		}

		page.SetDefaultTimeout(120000)
		page.SetDefaultNavigationTimeout(120000)

		result, err := test.Run(page, test.GetIntegrationTest())
		if err != nil {
			return testResultMsg{
				testIndex: j.testIndex,
				iteration: j.iteration,
				result:    errMsg{fmt.Errorf("test '%s' failed: %w", test.Name(), err)},
			}
		}

		return testResultMsg{testIndex: j.testIndex, iteration: j.iteration, result: result}
	}
}

func calculateAverageResults(results []*tests.IntegrationResult) string {
	if len(results) == 0 {
		return "No results yet"
	}

	totalDuration := time.Duration(0)
	totalSimilarity := 0.0
	successfulTests := 0
	testDurations := make(map[string]time.Duration)
	testSimilarities := make(map[string]float64)
	testCounts := make(map[string]int)

	for _, result := range results {
		totalDuration += result.Duration
		totalSimilarity += result.Similarity
		successfulTests++

		testDurations[result.TestName] += result.Duration
		testSimilarities[result.TestName] += result.Similarity
		testCounts[result.TestName]++
	}

	if successfulTests == 0 {
		return "No successful tests to average."
	}

	averageDuration := totalDuration / time.Duration(successfulTests)
	averageSimilarity := totalSimilarity / float64(successfulTests)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Total tests: %d, Successful tests: %d\n", len(results), successfulTests))
	result.WriteString(fmt.Sprintf("Overall Average Duration: %s, Overall Average Similarity: %.2f%%\n\n", averageDuration, averageSimilarity*100))

	result.WriteString("Averages per test:\n")
	for name, count := range testCounts {
		avgDuration := testDurations[name] / time.Duration(count)
		avgSimilarity := (testSimilarities[name] / float64(count)) * 100
		result.WriteString(fmt.Sprintf("- %s: Average Duration: %s, Average Similarity: %.2f%%\n", name, avgDuration, avgSimilarity))
	}

	return result.String()
}

// nolint:funlen,gocognit,cyclop
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}

	case startBenchmarkMsg:
		m.running = true
		m.results = nil
		m.errorMsg = ""
		pw, err := playwright.Run()
		if err != nil {
			m.errorMsg = err.Error()
			m.running = false
			return m, tea.Quit
		}
		m.pw = pw

		browser, err := m.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(true),
		})
		if err != nil {
			m.errorMsg = fmt.Sprintf("could not launch browser: %v", err)
			m.running = false
			return m, tea.Quit
		}
		m.browser = browser

		var jobList []job
		for testIdx := range tests.Tests {
			for i := 0; i < m.iterations; i++ {
				jobList = append(jobList, job{testIndex: testIdx, iteration: i})
			}
		}
		m.totalJobs = len(jobList)
		m.jobs = make(chan job, m.totalJobs)
		for _, j := range jobList {
			m.jobs <- j
		}
		close(m.jobs)

		var cmds []tea.Cmd
		for i := 0; i < m.parallelism; i++ {
			if j, ok := <-m.jobs; ok {
				cmds = append(cmds, runJobCmd(m.browser, j))
			}
		}
		m.currentAction = fmt.Sprintf("Starting %d workers...", len(cmds))
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 4
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

	case errMsg:
		m.running = false
		m.errorMsg = msg.err.Error()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.running && m.errorMsg == "" {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case testResultMsg:
		m.completedJobs++
		testName := tests.Tests[msg.testIndex].Name()

		switch r := msg.result.(type) {
		case errMsg:
			log.Printf("Job failed: test '%s', iteration %d, error: %v", testName, msg.iteration+1, r.err)
		case *tests.IntegrationResult:
			r.TestName = testName
			m.results = append(m.results, r)
			log.Printf("Job finished: Test: %s, Iteration: %d, Duration: %s, Similarity: %.2f%%",
				r.TestName, msg.iteration+1, r.Duration, r.Similarity*100)
			if m.debug {
				log.Printf("Actual Response for Test '%s', Iteration %d:\n%s", r.TestName, msg.iteration+1, r.ActualResponse)
			}
		}

		if m.completedJobs >= m.totalJobs {
			m.running = false
			log.Println("All jobs finished.")
			return m, nil
		}

		if j, ok := <-m.jobs; ok {
			m.currentAction = fmt.Sprintf("Running test: %s", tests.Tests[j.testIndex].Name())
			return m, runJobCmd(m.browser, j)
		}
		// No more jobs to start, just wait for others to finish
		m.currentAction = "Waiting for remaining jobs to finish..."
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	s := titleStyle.Render("Smart Benchmarking Tool") + "\n\n"

	if m.errorMsg != "" {
		return s + fmt.Sprintf("Error: %s\n\nPress any key to quit.", m.errorMsg)
	}

	if !m.running && len(m.results) == 0 {
		s += fmt.Sprintf("%s Initializing benchmark...", m.spinner.View())
		if m.debug {
			s += "\n\n" + m.currentAction
		}
		return s
	}

	if m.running {
		percent := float64(m.completedJobs) / float64(m.totalJobs)

		var recentResults strings.Builder
		if len(m.results) > 0 {
			start := 0
			if len(m.results) > 5 {
				start = len(m.results) - 5
			}
			for _, res := range m.results[start:] {
				recentResults.WriteString(fmt.Sprintf("Test: %s, Duration: %s, Similarity: %.2f%%\n", res.TestName, res.Duration, res.Similarity*100))
			}
		}

		debugInfo := ""
		if m.debug {
			debugInfo = "\n\n" + m.currentAction
		}

		runningView := fmt.Sprintf("%s Running benchmarks... (%d/%d)\n\n%s\n\n%s%s",
			m.spinner.View(),
			m.completedJobs,
			m.totalJobs,
			m.progress.ViewAs(percent),
			recentResults.String(),
			debugInfo,
		)
		return s + runningView
	}

	// Finished
	return s + "All tests completed!\n\n" + calculateAverageResults(m.results) + "\n\nPress q to quit."
}

func main() {
	debug := flag.Bool("debug", false, "show debug information")
	parallelism := flag.Int("p", defaultParallelism, "number of parallel tests to run")
	iterations := flag.Int("i", iterationsPerTest, "number of iterations per test")
	flag.Parse()

	const logfileName = "benchmark.log"
	// We remove the logfile on each start to ensure a clean log.
	// We ignore the error, as the file may not exist on the first run.
	_ = os.Remove(logfileName)

	f, err := tea.LogToFile(logfileName, "benchmark")
	if err != nil {
		fmt.Println("could not create log file:", err)
		os.Exit(1)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("could not close log file:", err)
			os.Exit(1)
		}
	}()
	log.Printf("Starting benchmark with: parallelism=%d, iterations=%d, debug=%t", *parallelism, *iterations, *debug)

	p := tea.NewProgram(initialModel(*debug, *parallelism, *iterations))
	m, err := p.Run()
	if finalModel, ok := m.(model); ok {
		if finalModel.browser != nil {
			if err := finalModel.browser.Close(); err != nil {
				log.Printf("could not close browser: %v", err)
			}
		}
		if finalModel.pw != nil {
			if err := finalModel.pw.Stop(); err != nil {
				log.Printf("could not stop Playwright: %v", err)
			}
		}
	}
	if err != nil {
		log.Printf("Alas, there's been an error: %v", err)
		os.Exit(1) //nolint:gocritic
	}
	// resets terminal when completing
	cmd := exec.Command("reset")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}
