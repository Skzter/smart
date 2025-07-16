package handler

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	shared "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

func TestNewAutoTesterController(t *testing.T) {
	tests := []struct {
		testName      string
		logger        *slog.Logger
		config        *config.Config
		service       service.OpenAIService
		expectedError bool
	}{
		{
			testName: "valid Service",
			logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			config: &config.Config{
				Model: "gpt-4",
				Port:  "8081",
				Prompts: &config.Prompts{
					ValidationPrompt:     "Bitte überprüfe die Eingabe auf Vollständigkeit.",
					AutoPlaywrightPrompt: "Erstelle automatisch ein Playwright-Skript für den folgenden Use Case.",
				},
				Timeout: 30,
				Region:  "us-central1",
				Bucket:  "my-app-bucket",
			},
			service:       &mocks.MockOpenAIService{},
			expectedError: false,
		}, {
			testName: "logger nil",
			logger:   nil,
			config: &config.Config{
				Model: "gpt-4",
				Port:  "8081",
				Prompts: &config.Prompts{
					ValidationPrompt:     "Bitte überprüfe die Eingabe auf Vollständigkeit.",
					AutoPlaywrightPrompt: "Erstelle automatisch ein Playwright-Skript für den folgenden Use Case.",
				},
				Timeout: 30,
				Region:  "us-central1",
				Bucket:  "my-app-bucket",
			},
			service:       &mocks.MockOpenAIService{},
			expectedError: true,
		}, {
			testName:      "config nil",
			logger:        slog.New(slog.NewTextHandler(os.Stdout, nil)),
			config:        nil,
			expectedError: true,
			service:       &mocks.MockOpenAIService{},
		}, {
			testName: "service nil",
			logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			config: &config.Config{
				Model: "gpt-4",
				Port:  "8081",
				Prompts: &config.Prompts{
					ValidationPrompt:     "Bitte überprüfe die Eingabe auf Vollständigkeit.",
					AutoPlaywrightPrompt: "Erstelle automatisch ein Playwright-Skript für den folgenden Use Case.",
				},
				Timeout: 30,
				Region:  "us-central1",
				Bucket:  "my-app-bucket",
			},
			service:       nil,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			controller, err := NewAutotesterController(test.logger, test.config, test.service)

			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil for test case: %s", test.testName)
				}
				if controller != nil {
					t.Errorf("Expected nil controller, got %+v", controller)
				}
			} else if !test.expectedError {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if controller == nil {
					t.Errorf("Expected controller instance, got nil")
				}
			}
		})
	}
}

// nolint:funlen
func TestHandleChatRequest(t *testing.T) {
	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		config         *config.Config
	}{
		{
			TestName: "Valid JSON",
			RequestBody: `{
				"message": {
					"data":  "Hello, how can I help you?",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusOK,
			config: &config.Config{
				Model:    "gpt-4.1-nano-2025-04-14",
				Port:     "8081",
				LogLevel: "info",
				Timeout:  20,
				Region:   "eu-central-1",
				Bucket:   "smart-autotester",
				Prompts: &config.Prompts{
					ValidationPrompt: `You are a helpful Assistant.
				Answer all questions precisely and unambiguously.
				For now, only answer with 'yes'.`,
					AutoPlaywrightPrompt: "",
				},
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":it ad json}`,
			ExpectedStatus: http.StatusBadRequest,
			config: &config.Config{
				Model:    "gpt-4.1-nano-2025-04-14",
				Port:     "8081",
				LogLevel: "info",
				Timeout:  20,
				Region:   "eu-central-1",
				Bucket:   "smart-autotester",
				Prompts: &config.Prompts{
					ValidationPrompt: `You are a helpful Assistant.
				Answer all questions precisely and unambiguously.
				For now, only answer with 'yes'.`,
					AutoPlaywrightPrompt: "",
				},
			},
		}, {
			TestName: "failed servicehandler",
			RequestBody: `{
				"message": {
					"data":  "Hello, how can I help you?",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			config: &config.Config{
				Model:    "gpt-4.1-nano-2025-04-14",
				Port:     "8081",
				LogLevel: "info",
				Timeout:  20,
				Region:   "eu-central-1",
				Bucket:   "smart-autotester",
				Prompts: &config.Prompts{
					ValidationPrompt: `You are a helpful Assistant.
				Answer all questions precisely and unambiguously.
				For now, only answer with 'yes'.`,
					AutoPlaywrightPrompt: "",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Errors.Errors()

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			mockOpenAiService := &mocks.MockOpenAIService{}

			if test.ExpectedStatus == http.StatusOK {
				mockOpenAiService.On("Request", ctx, mock.Anything).Return(&shared.Response{
					Text:      "Richtig",
					SessionID: "1234",
				}, nil)
			} else {
				mockOpenAiService.On("Request", ctx, mock.Anything).Return(nil, fmt.Errorf("MockError"))
			}

			controller, err := NewAutotesterController(logger, test.config, mockOpenAiService)

			if err != nil {
				t.Errorf("build failed")
			}
			controller.HandleChatRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleUserInfoRequest(t *testing.T) {
	tests := []struct {
		TestName        string
		UserRequestBody string
		ResponseForUser entity.ResponseForUser
		Context         context.Context
		ExpectedStatus  int
	}{
		{
			TestName: "Valid UserRequestBody",
			UserRequestBody: `{
				"userId": "9177b856-46a0-11f0-9fe2-0242ac120002",
				"allConversations": [
				  {
					"ConversationId": "string",
					"Messages": [
					  {
						"data": "string",
						"agent": "user"
					  }
					]
				  }
				]
			  }`,
			Context:        context.Background(),
			ExpectedStatus: http.StatusOK,
		}, {
			TestName:        "Invalid UserRequestBody",
			UserRequestBody: ``,
			Context:         context.Background(),
			ExpectedStatus:  http.StatusBadRequest,
		},
	}

	config := &config.Config{
		Model:    "gpt-4.1-nano-2025-04-14",
		Port:     "8081",
		LogLevel: "info",
		Timeout:  20,
		Region:   "eu-central-1",
		Bucket:   "smart-autotester",
		Prompts: &config.Prompts{
			ValidationPrompt: `You are a helpful Assistant.
		Answer all questions precisely and unambiguously.
		For now, only answer with 'yes'.`,
			AutoPlaywrightPrompt: "",
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/v1/userInfo", bytes.NewBufferString(test.UserRequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Errors.Errors()

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			mockOpenAiService := &mocks.MockOpenAIService{}

			if test.ExpectedStatus == http.StatusOK {
				mockOpenAiService.On("Request", ctx, mock.Anything).Return(&shared.Response{
					Text:      "Richtig",
					SessionID: "1234",
				}, nil)
			} else {
				mockOpenAiService.On("Request", ctx, mock.Anything).Return(nil, fmt.Errorf("MockError"))
			}

			controller, err := NewAutotesterController(logger, config, mockOpenAiService)

			if err != nil {
				t.Errorf("build failed")
			}
			controller.HandleUserInfoRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}
