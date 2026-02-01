# Level 4: LLM Integration Flow

This document describes the detailed code-level flow for LLM-powered test generation in S.M.A.R.T.

## Diagram

![LLM Flow](diagrams/llm-flow.mmd.svg)
See [llm-flow.mmd](diagrams/llm-flow.mmd) for the Mermaid source. The SVG is generated from it (see [Regenerating SVGs](../README.md#regenerating-svgs) in the architecture README).

## Overview

The LLM integration flow enables natural language test generation by orchestrating interactions between the frontend, Autotester service, and LLM API. Validation and generation are always combined (validate then generate); user prompt goes directly to validation then to Generate. There is no separate Prompt Builder; the diagram's LLM Test Suite (which tests the system prompt) is not part of the regular LLM flow.

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

4. **Chat Manager**
   - ChatManager creates or orchestrates loading/updating chat (chat entity only; no orchestration of LLM flow)
   - Retrieves chat history from ChatStorageService
   - User prompt goes directly to validation then to Generate (no Prompt Builder)

5. **Validation and Generation**
   - Validation and generate flow are always combined (not separate)
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
       "model": "gpt-5-mini-2025-08-07",
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
   - Response is pass-through of the LLM JSON response: either test code only, or (if prompt validation failed) only hints for the prompt
   - No separate code extraction/validation step in this flow

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

### 2. Validation and Generation (combined)

**Purpose:** Validation and generation are always combined: user prompt goes to validation then to Generate (no separate validation-only flow in practice). Frontend calls `/validate` then `/chat`. Response is either test code or (if validation failed) hints for the prompt.

---

## Key Components and Their Roles

### ChatManager
**File:** `internal/autotester/domain/service/chatManager.go`

**Responsibilities:**
- Chat entity management (create, save, update chats)
- Manage conversation context
- Does not orchestrate LLM flow; validation and generation are handled by other services

**Key Methods:**
- `ProcessChat(request)` - Main entry point
- `BuildContext()` - Gather chat history
- `HandleResponse()` - Process LLM output (pass-through of JSON: test or validation hints)

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

**Note:** Prompt Builder is not used; user prompt goes directly to validation then to Generate. Response processing is pass-through of the LLM JSON (test code or validation hints).

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

*(All items below are future/planned if not already implemented.)*

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

*Code examples should be verified against the current codebase; implementation details may differ.*

### LLM API Call (conceptual)
```go
func (s *LLMTestSuite) CallLLM(prompt string) (string, error) {
    req := openai.ChatCompletionRequest{
        Model: "gpt-5-mini-2025-08-07",
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
  model = "gpt-5-mini-2025-08-07"
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
