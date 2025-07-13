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
	iterationsPerTest = 100
)

// nolint:gochecknoglobals
var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)
)

type errMsg struct{ err error }

type startBenchmarkMsg struct{}

type model struct {
	pw               *playwright.Playwright
	running          bool
	spinner          spinner.Model
	progress         progress.Model
	results          []string
	currentTest      int
	currentIteration int
	errorMsg         string
	debug            bool
	currentAction    string
}

func initialModel(debug bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner:  s,
		progress: progress.New(progress.WithDefaultGradient()),
		debug:    debug,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.EnterAltScreen,
		func() tea.Msg { return startBenchmarkMsg{} },
	)
}

func runSingleTest(pw *playwright.Playwright, test tests.BenchmarkTest) tea.Cmd {
	return func() tea.Msg {
		browser, err := pw.Chromium.Launch()
		if err != nil {
			return errMsg{fmt.Errorf("could not launch browser: %w", err)}
		}
		defer func() {
			if err := browser.Close(); err != nil {
				log.Printf("could not close browser: %v", err)
			}
		}()

		page, err := browser.NewPage()
		if err != nil {
			return errMsg{fmt.Errorf("could not create page: %w", err)}
		}

		result, err := test.Run(page)
		if err != nil {
			return errMsg{fmt.Errorf("test '%s' failed: %w", test.Name(), err)}
		}

		return result
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

func (m model) nextIteration() (model, tea.Cmd) {
	m.currentIteration++
	if m.currentIteration >= iterationsPerTest {
		m.currentIteration = 0
		m.currentTest++
	}

	if m.currentTest >= len(tests.Tests) {
		m.running = false
		return m, nil
	}

	m.currentAction = fmt.Sprintf("Running test: %s", tests.Tests[m.currentTest].Name())
	return m, runSingleTest(m.pw, tests.Tests[m.currentTest])
}

// nolint:funlen
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg == nil {
		if m.running {
			testName := tests.Tests[m.currentTest].Name()
			m.results = append(m.results, fmt.Sprintf("Test: %s, Iteration: %d, Completed", testName, m.currentIteration+1))
			return m.nextIteration()
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}

	case startBenchmarkMsg:
		m.running = true
		m.currentTest = 0
		m.currentIteration = 0
		m.results = nil
		m.errorMsg = ""
		pw, err := playwright.Run()
		if err != nil {
			m.errorMsg = err.Error()
			m.running = false
			return m, nil
		}
		m.pw = pw
		m.currentAction = fmt.Sprintf("Running test: %s", tests.Tests[m.currentTest].Name())
		return m, runSingleTest(m.pw, tests.Tests[m.currentTest])

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
		if (m.running || len(m.results) == 0) && m.errorMsg == "" {
			m.spinner, cmd = m.spinner.Update(msg)
		}
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case *tests.LoginResult:
		testName := tests.Tests[m.currentTest].Name()
		m.results = append(m.results, fmt.Sprintf("Test: %s, Iteration: %d, Duration: %s", testName, m.currentIteration+1, msg.Duration))
		return m.nextIteration()

	case *tests.ExampleResult:
		testName := tests.Tests[m.currentTest].Name()
		m.results = append(m.results, fmt.Sprintf("Test: %s, Iteration: %d, Duration: %s, Similarity: %.2f%%", testName, m.currentIteration+1, msg.Duration, msg.Similarity*100))
		return m.nextIteration()
	}

	return m, nil
}

func (m model) View() string {
	s := titleStyle.Render("Smart Benchmarking Tool") + "\n\n"

	if m.errorMsg != "" {
		return s + fmt.Sprintf("Error: %s\n\nPress any key to quit.", m.errorMsg)
	}

	if !m.running && len(m.results) == 0 {
		return s + fmt.Sprintf("%s Initializing benchmark...", m.spinner.View())
	}

	if m.running {
		totalTests := len(tests.Tests)
		totalIterations := totalTests * iterationsPerTest
		completedIterations := m.currentTest*iterationsPerTest + m.currentIteration

		percent := float64(completedIterations) / float64(totalIterations)

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

		runningView := fmt.Sprintf("%s Running test %d/%d (Iteration %d/%d)\n\n%s\n\n%s%s",
			m.spinner.View(),
			m.currentTest+1,
			totalTests,
			m.currentIteration+1,
			iterationsPerTest,
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
	flag.Parse()

	if *debug {
		f, err := tea.LogToFile("debug.log", "debug")
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
	}

	p := tea.NewProgram(initialModel(*debug))
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
