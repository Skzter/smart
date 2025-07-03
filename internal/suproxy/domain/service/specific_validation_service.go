package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// Template for the system message sent to OpenAI
// %s will be replaced by the question from the user (in main.go)
const specificPromptTemplate = `
You are a mock data tester throughout this entire conversation.
In each case, I will provide a request file and the corresponding response file.
Your task is to read both files and then answer the respective question based solely on their content.
Your response must be:

brief, precise, and clearly structured
entirely in valid JSON format
without any additional explanation, commentary, or formatting
This instruction applies to all further interactions in this chat. Once the request and response files are provided, you will always follow this behavior without the need to repeat the task description. 
The current question is: "%s"
`

// SpecificValidationService is a struct that contains the OpenAI client
// It handles the validation logic for one specific question
type SpecificValidationService struct {
	client *openai.Client
}

// NewSpecificValidationService creates a new instance of the validation service
// It takes an OpenAI API key and returns a ready-to-use service
func NewSpecificValidationService(apiKey string) *SpecificValidationService {
	client := openai.NewClient(apiKey)
	return &SpecificValidationService{client: client}
}

// ValidateRequestResponse loads a request and response file,
// sends both to the OpenAI API together with a question,
// and returns the answer as a JSON object
func (s *SpecificValidationService) ValidateRequestResponse(
	requestPath string,
	responsePath string,
	question string,
) (map[string]interface{}, error) {

	// Read the contents of the request file
	requestData, err := os.ReadFile(requestPath)
	if err != nil {
		return nil, fmt.Errorf("could not read request file: %w", err)
	}

	// Read the contents of the response file
	responseData, err := os.ReadFile(responsePath)
	if err != nil {
		return nil, fmt.Errorf("could not read response file: %w", err)
	}

	// Combine the request and response into one user message
	userMessage := fmt.Sprintf("Request:\n%s\n\nResponse:\n%s", string(requestData), string(responseData))

	// Fill in the prompt template with the actual question
	systemMessage := fmt.Sprintf(specificPromptTemplate, question)

	// Send everything to the OpenAI API
	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o, // use GPT-4o model
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMessage, // prompt instruction
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage, // the input data
				},
			},
			Temperature: 0.0, // deterministic output
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Try to parse the response content as JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w\n\nRaw Output:\n%s", err, resp.Choices[0].Message.Content)
	}

	// Return the parsed result
	return result, nil
}
