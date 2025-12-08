package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	PromptPath  = "configs/prompts/autoplaywright_prompt.txt"
	Iterations  = 5
	OpenAIModel = "gpt-4o" // Use a strong model for optimization
)

func main() {
	apiKey := os.Getenv("OPENAI_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_KEY environment variable is required")
	}

	log.Println("Starting Prompt Optimization Loop...")

	// Initialize Eino Chat Model for the Optimizer Agent
	chatModel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey: apiKey,
		Model:  OpenAIModel,
	})
	if err != nil {
		log.Fatalf("Failed to create chat model: %v", err)
	}

	// Load Initial Prompt
	currentPromptBytes, err := os.ReadFile(PromptPath)
	if err != nil {
		log.Fatalf("Failed to read initial prompt: %v", err)
	}
	currentPrompt := string(currentPromptBytes)

	for i := 1; i <= Iterations; i++ {
		log.Printf("=== Optimization Round %d/%d ===\n", i, Iterations)

		// 1. Run Benchmark
		log.Println("Running benchmark...")
		successRate, failureLog := runBenchmark()
		log.Printf("Benchmark Result: Success Rate: %.2f%%\n", successRate)

		if successRate >= 100.0 {
			log.Println("Perfect score achieved! Stopping optimization.")
			break
		}

		// 2. Optimization Step
		log.Println("Optimizing prompt based on failures...")
		newPrompt, err := optimizePrompt(context.Background(), chatModel, currentPrompt, failureLog)
		if err != nil {
			log.Printf("Optimization failed: %v\n", err)
			continue
		}

		// 3. Update Prompt File
		currentPrompt = newPrompt
		if err := os.WriteFile(PromptPath, []byte(currentPrompt), 0644); err != nil {
			log.Fatalf("Failed to write updated prompt: %v", err)
		}
		log.Println("Prompt updated.")
	}
}

// runBenchmark executes the benchmark tool and captures the output.
// It returns the success rate (0-100) and a string containing failure logs.
func runBenchmark() (float64, string) {
	// We run the benchmark in a subprocess.
	// Note: We need to modify cmd/benchmark to support a "headless" or "report" mode
	// that outputs machine-parsable results.
	// For now, let's assume standard output parsing or passing a flag like "-report json".

	// Since we haven't updated benchmark yet, let's assume we can grep the logs or output.
	// Ideally, `cmd/benchmark` should write a report file.

	cmd := exec.Command("go", "run", "cmd/benchmark/main.go", "-headless", "-iterations", "3")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		log.Printf("Benchmark execution failed (might be test failures): %v\n", err)
	}

	// Parse output to find success rate and failures.
	// This is a naive parser based on expected log output.
	// "Success Rate: 85.00%"
	// "Job failed: test 'TestName', error: ..."

	// Extract Success Rate
	successRate := 0.0
	// Regex to find "Success Rate: XX.XX%" - purely hypothetical matching logic
	// We will update benchmark to output this explicitly.
	// For now, let's look for "Job finished" vs "Job failed" counts

	finishedCount := strings.Count(outputStr, "Job finished:")
	failedCount := strings.Count(outputStr, "Job failed:")
	total := finishedCount + failedCount

	if total > 0 {
		successRate = (float64(finishedCount) / float64(total)) * 100.0
	}

	// Extract Failures
	// Simple filter for lines containing "Job failed"
	var failures strings.Builder
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Job failed") || strings.Contains(line, "error:") {
			failures.WriteString(line + "\n")
		}
	}

	return successRate, failures.String()
}

func optimizePrompt(ctx context.Context, model model.ChatModel, currentPrompt, failureLogs string) (string, error) {
	// Construct the Meta-Prompt
	systemInstruction := `You are an expert Prompt Engineer for LLM-based testing agents.
Your task is to analyze the failures of an AutoPlaywright agent and refine its System Prompt to prevent these errors in the future.
- You must keep the core structure and constraints of the prompt intact.
- You should clarify instructions, add examples, or tighten constraints based on the specific errors provided.
- Return ONLY the updated System Prompt text. No markdown, no conversational filler.`

	userMessage := fmt.Sprintf(`Current System Prompt:
"""
%s
"""

Benchmark Failures:
"""
%s
"""

Please rewrite the System Prompt to fix these issues.`, currentPrompt, failureLogs)

	// Build Eino Chain
	// Simple Chain: Input -> Model -> Output
	chain := compose.NewChain[[]*schema.Message, *schema.Message]()
	chain.AppendChatModel(model)

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return "", err
	}

	result, err := runnable.Invoke(ctx, []*schema.Message{
		schema.SystemMessage(systemInstruction),
		schema.UserMessage(userMessage),
	})
	if err != nil {
		return "", err
	}

	return result.Content, nil
}
