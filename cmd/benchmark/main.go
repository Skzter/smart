package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
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
	iterationsPerTest  = 100
	defaultParallelism = 4
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
	running       bool
	spinner       spinner.Model
	progress      progress.Model
	results       []string
	errorMsg      string
	debug         bool
	currentAction string
	parallelism   int
	jobs          chan job
	totalJobs     int
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

func runJobCmd(pw *playwright.Playwright, j job) tea.Cmd {
	return func() tea.Msg {
		test := tests.Tests[j.testIndex]
		log.Printf("Starting job: test '%s', iteration %d", test.Name(), j.iteration+1)
		browser, err := pw.Chromium.Launch()
		if err != nil {
			return testResultMsg{
				testIndex: j.testIndex,
				iteration: j.iteration,
				result:    errMsg{fmt.Errorf("could not launch browser: %w", err)},
			}
		}
		defer func() {
			if err := browser.Close(); err != nil {
				log.Printf("could not close browser: %v", err)
			}
		}()

		page, err := browser.NewPage()
		if err != nil {
			return testResultMsg{
				testIndex: j.testIndex,
				iteration: j.iteration,
				result:    errMsg{fmt.Errorf("could not create page: %w", err)},
			}
		}

		result, err := test.Run(page)
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

func calculateAverageResults(results []string) string {
	if len(results) == 0 {
		return "No results yet"
	}

	totalDuration := time.Duration(0)
	totalSimilarity := 0.0
	totalIterations := 0
	hasSimilarity := false

	for _, result := range results {
		parts := strings.Split(result, ", ")
		var duration time.Duration
		var similarity float64
		var err error

		// Try to find duration and similarity in the result string
		for _, part := range parts {
			if strings.HasPrefix(part, "Duration: ") {
				durationStr := strings.TrimPrefix(part, "Duration: ")
				duration, err = time.ParseDuration(durationStr)
				if err != nil {
					continue
				}
				totalDuration += duration
			} else if strings.HasPrefix(part, "Similarity: ") {
				simStr := strings.TrimPrefix(part, "Similarity: ")
				simStr = strings.TrimSuffix(simStr, "%")
				simStr = strings.TrimSpace(simStr)
				sim, err := fmt.Sscanf(simStr, "%f", &similarity)
				if err != nil || sim != 1 {
					// fallback: try ParseFloat
					similarity, err = strconv.ParseFloat(simStr, 64)
					if err != nil {
						continue
					}
				}
				totalSimilarity += similarity
				hasSimilarity = true
			}
		}
		totalIterations++
	}

	if totalIterations == 0 {
		return "No valid results to average"
	}

	averageDuration := totalDuration / time.Duration(totalIterations)
	var averageSimilarity float64
	if hasSimilarity {
		averageSimilarity = totalSimilarity / float64(totalIterations)
		return fmt.Sprintf("Average duration: %s, Average similarity: %.2f%%", averageDuration, averageSimilarity)
	}
	return fmt.Sprintf("Average duration: %s", averageDuration)
}

// nolint:funlen
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
				cmds = append(cmds, runJobCmd(m.pw, j))
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
		var newResult string
		testName := tests.Tests[msg.testIndex].Name()

		switch r := msg.result.(type) {
		case errMsg:
			m.running = false
			m.errorMsg = r.err.Error()
			log.Printf("Job failed: test '%s', iteration %d, error: %v", testName, msg.iteration+1, r.err)
			return m, nil
		case *tests.LoginResult:
			newResult = fmt.Sprintf("Test: %s, Iteration: %d, Duration: %s", testName, msg.iteration+1, r.Duration)
		case *tests.ExampleResult:
			newResult = fmt.Sprintf("Test: %s, Iteration: %d, Duration: %s, Similarity: %.2f%%", testName, msg.iteration+1, r.Duration, r.Similarity*100)
		}
		m.results = append(m.results, newResult)
		log.Printf("Job finished: %s", newResult)

		if len(m.results) == m.totalJobs {
			m.running = false
			log.Println("All jobs finished.")
			return m, nil
		}

		if j, ok := <-m.jobs; ok {
			m.currentAction = fmt.Sprintf("Running test: %s", tests.Tests[j.testIndex].Name())
			return m, runJobCmd(m.pw, j)
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
		percent := float64(len(m.results)) / float64(m.totalJobs)

		var recentResults string
		if len(m.results) > 0 {
			start := 0
			if len(m.results) > 5 {
				start = len(m.results) - 5
			}
			recentResults = strings.Join(m.results[start:], "\n")
		}

		debugInfo := ""
		if m.debug {
			debugInfo = "\n\n" + m.currentAction
		}

		runningView := fmt.Sprintf("%s Running benchmarks... (%d/%d)\n\n%s\n\n%s%s",
			m.spinner.View(),
			len(m.results),
			m.totalJobs,
			m.progress.ViewAs(percent),
			recentResults,
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

	f, err := tea.LogToFile("benchmark.log", "benchmark")
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
	if finalModel, ok := m.(model); ok && finalModel.pw != nil {
		if err := finalModel.pw.Stop(); err != nil {
			log.Printf("could not stop Playwright: %v", err)
		}
	}
	if err != nil {
		log.Printf("Alas, there's been an error: %v", err)
		os.Exit(1) //nolint:gocritic
	}
}
