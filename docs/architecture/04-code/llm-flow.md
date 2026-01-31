# Level 4: LLM Integration Flow

This document describes the detailed code-level flow for LLM-powered test generation in S.M.A.R.T.

## Diagram

![LLM Flow](diagrams/llm-flow.mmd.svg)
See [llm-flow.mmd](diagrams/llm-flow.mmd) for the Mermaid source.

## Overview

The LLM integration flow enables natural language test generation by orchestrating interactions between the frontend, Autotester service, and LLM API. The flow supports both validation and generation modes.

## Sequence Diagrams

### 1. Test Generation Flow (Complete)

**Actors:**
- User (Developer/Tester)
- Frontend (Svelte)
- Autotester API
- Chat Manager Service
- LLM Test Suite Service
- LLM API (OpenAI/Claude)
- Chat Storage
- S3 / Object Storage (Parquet)

**Steps:**

1. **User Input**
   - User types test description in chat interface
   - Example: "Create a test that checks if the login button is visible"

2. **Frontend Processing**
   - Chat component captures input
   - API service constructs ChatRequest:
     ```typescript
     {
       message: { body: "Create a test..." },
       userId: "user-123",
       chatId: "chat-456"
     }
     ```
   - POST to `/api/v1/chat`

3. **Autotester API Layer**
   - Router receives request at `/api/v1/chat`
   - Routes to AutotesterController.HandleChatRequest()
   - Validates request structure
   - Extracts userId, chatId, message

4. **Chat Manager Orchestration**
   - ChatManager.ProcessChat() called
   - Retrieves chat history from ChatStorageService
   - Builds context from previous messages
   - Determines if validation or generation needed

5. **Prompt Construction**
   - LLMTestSuite.BuildPrompt() called
   - Loads prompt template from Pkl config
   - Constructs system prompt with:
     - Test generation instructions
     - Playwright syntax guidelines
     - Best practices
     - Template structure
   - Adds user context:
     - Previous conversation
     - Current request
     - Available tools/selectors

6. **LLM API Call**
   - LLMTestSuite.CallLLM() invoked
   - Prepares request:
     ```json
     {
       "model": "gpt-5.2",
       "messages": [
         {"role": "system", "content": "You are..."},
         {"role": "user", "content": "Create a test..."}
       ],
       "temperature": 0.7
     }
     ```
   - Makes HTTP POST to OpenAI/Claude API
   - Handles streaming response (optional)

7. **Response Processing**
   - LLMTestSuite.ParseResponse() called
   - Extracts test code from response
   - Validates code structure
   - Checks for:
     - Valid Playwright syntax
     - Proper async/await usage
     - Required imports
     - Test structure

8. **Code Extraction**
   - Extracts code blocks from markdown
   - Identifies language (TypeScript/JavaScript)
   - Removes markdown formatting
   - Validates completeness

9. **Persistence**
   - ChatStorageService.SaveMessage() called
   - Stores user message in S3 (Parquet)
   - Stores LLM response in S3 (Parquet)
   - Updates chat metadata (timestamp, message count)

10. **Response Construction**
    - AutotesterController builds response:
      ```json
      {
        "message": {
          "body": "```typescript\n// Generated test code\n```"
        },
        "chatId": "chat-456",
        "userId": "user-123"
      }
      ```
    - Returns HTTP 200 with response

11. **Frontend Display**
    - API service receives response
    - Updates shared state with new message
    - Chat component re-renders
    - Syntax highlighting applied to code
    - User sees generated test

---

### 2. Validation Flow

**Purpose:** Validate user prompt before full generation (faster feedback)

**Steps:**

1. **User Input**
   - User types prompt
   - Frontend may trigger validation on-the-fly

2. **Frontend Processing**
   - API service constructs ChatRequest
   - POST to `/api/v1/validate`

3. **Autotester API Layer**
   - Router routes to AutotesterController.HandleChatRequestValidity()
   - Validates request

4. **Validator Service**
   - Validator.ValidatePrompt() called
   - Checks prompt structure
   - Verifies clarity and completeness
   - May call LLM for semantic validation

5. **LLM Validation Call**
   - Simplified prompt to LLM:
     - "Is this test request clear and actionable?"
     - Returns yes/no with suggestions
   - Faster, cheaper than full generation

6. **Validation Response**
   - Returns validation result:
     ```json
     {
       "message": {
         "body": "Prompt is valid. Ready to generate test."
       },
       "isValid": true,
       "suggestions": []
     }
     ```

7. **Frontend Feedback**
   - Displays validation result
   - Shows suggestions if needed
   - Enables/disables generate button

---

## Key Components and Their Roles

### ChatManager
**File:** `internal/autotester/domain/service/chatManager.go`

**Responsibilities:**
- Orchestrate entire chat flow
- Manage conversation context
- Coordinate between services

**Key Methods:**
- `ProcessChat(request)` - Main entry point
- `BuildContext()` - Gather chat history
- `HandleResponse()` - Process LLM output

### LLMTestSuite
**File:** `internal/autotester/domain/service/llmTestSuite.go`

**Responsibilities:**
- Prompt engineering
- LLM API communication
- Response parsing

**Key Methods:**
- `BuildPrompt(userMessage, context)` - Construct prompt
- `CallLLM(prompt)` - API call to LLM
- `ParseResponse(llmOutput)` - Extract test code
- `ValidateTestCode(code)` - Syntax validation

### Prompt Builder
**Responsibilities:**
- Load templates from Pkl config
- Inject user context
- Format for LLM consumption

**Template Structure:**
```
System: You are an expert test automation engineer...
- Use Playwright for browser automation
- Write tests in TypeScript
- Follow best practices...

User Context:
- Previous tests: [...]
- Current request: [user message]

Task: Generate a Playwright test for: [specific requirement]
```

### Response Parser
**Responsibilities:**
- Extract code from markdown
- Validate syntax
- Clean formatting

**Parsing Logic:**
1. Identify code blocks (```typescript...```)
2. Extract code content
3. Remove comments if excessive
4. Validate structure
5. Return cleaned code

---

## Error Handling

### Frontend Errors
- Network failures: Retry with exponential backoff
- Invalid responses: Display error message
- Timeout: Cancel request, notify user

### Backend Errors
- LLM API failures: Retry with backoff
- Rate limiting: Queue requests
- Invalid prompts: Return validation error
- Parse failures: Request regeneration

### LLM API Errors
- 429 (Rate Limit): Exponential backoff
- 500 (Server Error): Retry with different model
- 400 (Bad Request): Fix prompt structure
- Timeout: Shorter timeout, retry

---

## Performance Optimizations

1. **Caching**
   - Cache common prompts
   - Redis for chat history
   - Template caching

2. **Streaming**
   - Stream LLM responses (SSE)
   - Progressive UI updates
   - Better perceived performance

3. **Prompt Optimization**
   - Minimal context (recent messages only)
   - Efficient token usage
   - Clear, concise instructions

4. **Parallel Processing**
   - Validation + generation in parallel (if needed)
   - Async S3 writes (Parquet)
   - Non-blocking LLM calls

---

## Code Examples

### Prompt Construction
```go
func (s *LLMTestSuite) BuildPrompt(userMessage string, chatHistory []Message) string {
    template := s.config.PromptTemplate
    
    // Build context from history
    context := formatChatHistory(chatHistory)
    
    // Inject into template
    prompt := fmt.Sprintf(template, context, userMessage)
    
    return prompt
}
```

### LLM API Call
```go
func (s *LLMTestSuite) CallLLM(prompt string) (string, error) {
    req := openai.ChatCompletionRequest{
        Model: "gpt-4",
        Messages: []openai.ChatCompletionMessage{
            {Role: "system", Content: systemPrompt},
            {Role: "user", Content: prompt},
        },
        Temperature: 0.7,
    }
    
    resp, err := s.client.CreateChatCompletion(ctx, req)
    if err != nil {
        return "", err
    }
    
    return resp.Choices[0].Message.Content, nil
}
```

### Code Extraction
```go
func (s *LLMTestSuite) ExtractCode(response string) (string, error) {
    // Find code blocks
    re := regexp.MustCompile("```(?:typescript|javascript)\\n([\\s\\S]*?)```")
    matches := re.FindStringSubmatch(response)
    
    if len(matches) < 2 {
        return "", ErrNoCodeFound
    }
    
    code := matches[1]
    
    // Validate syntax
    if err := s.ValidateCode(code); err != nil {
        return "", err
    }
    
    return code, nil
}
```

---

## Configuration

### Pkl Configuration
**File:** `configs/autotester.pkl`

```pkl
prompts {
  systemPrompt = """
    You are an expert test automation engineer...
  """
  
  testTemplate = """
    import { test, expect } from '@playwright/test';
    
    test('{{testName}}', async ({ page }) => {
      // Test implementation
    });
  """
}

llm {
  provider = "openai"
  model = "gpt-4"
  temperature = 0.7
  maxTokens = 2000
  timeout = "30s"
}
```

---

## Monitoring and Observability

### Metrics
- LLM API latency
- Token usage per request
- Success/failure rates
- Prompt length distribution

### Tracing
- Datadog APM traces entire flow
- Span for each component
- Identifies bottlenecks

### Logging
- Log prompts (sanitized)
- Log responses (truncated)
- Log errors with context
- Performance metrics
