package handler

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service/mocks"
)

func TestNewAutoTesterController(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		testName      string
		logger        *slog.Logger
		config        *config.Config
		expectedError bool
	}{
		{
			testName:      "valid parameters",
			logger:        logger,
			config:        cfg,
			expectedError: false,
		}, {
			testName:      "invalid parameters => logger nil",
			logger:        nil,
			config:        cfg,
			expectedError: true,
		},
	}

	// only two version because we shouldnt be testing the functionality of the assert
	// if it works once, it should work all the time
	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidatePrompt(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			controller, err := NewAutotesterController(test.logger, test.config, mockValServ, mockGenServ, mockLocalStorageServ)

			if test.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil for test case: %s", test.testName)
				}
				if controller != nil {
					t.Errorf("Expected nil controller, got %+v", controller)
				}
			} else {
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
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	// context with mock.Anything
	validPrompt := "this is a valid prompt"
	invalidPrompt := "this is a invalid prompt"
	sessionid := "2"

	mockSetup := []struct {
		function         string
		userPrompt       string
		sessionID        string
		expectedResponse string
		ResponseError    error
		expectedError    bool
	}{
		{
			function:         "ValidatePrompt",
			userPrompt:       validPrompt,
			sessionID:        sessionid,
			expectedResponse: "true",
			ResponseError:    nil,
			expectedError:    false,
		},
		{
			function:         "ValidatePrompt",
			userPrompt:       invalidPrompt,
			sessionID:        sessionid,
			expectedResponse: "false",
			ResponseError:    errors.New("prompt does not contain required information for test generation"),
			expectedError:    true,
		},
		{
			function:         "GeneratePrompt",
			userPrompt:       validPrompt,
			sessionID:        sessionid,
			expectedResponse: "This is a generated Prompt",
			ResponseError:    nil,
			expectedError:    false,
		},
	}
	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
	}{
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":it ad json}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		// testing correct requests
		{
			TestName: "valid request",
			RequestBody: `{
				"message": {
					"data":  "this is a valid prompt",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusOK,
		},
		{
			TestName: "invalid request => invalid prompt",
			RequestBody: `{
				"message": {
					"data":  "this is a invalid prompt",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidatePrompt(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)

	// setup mocks
	for _, mc := range mockSetup {
		if mc.function == "ValidatePrompt" {
			mockValServ.On(mc.function, mock.Anything, mc.userPrompt, mc.sessionID).Return(mc.ResponseError)
		}
		if mc.function == "GeneratePrompt" {
			mockGenServ.On(mc.function, mock.Anything, mc.userPrompt, mc.sessionID).Return(mc.expectedResponse, mc.ResponseError)
		}
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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ)

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
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
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

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidatePrompt(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)

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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ)

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

// nolint:dupl
func TestHandleSaveLocalRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		SetupMock      func(*mocks.MockTestcaseLocalStorageService)
	}{
		{
			TestName: "Valid save request",
			RequestBody: `{
				"userId": "user123",
				"conversationId": "conv456",
				"code": "import { test, expect } from '@playwright/test';\n\ntest('example test', async ({ page }) => {\n  await page.goto('https://example.com');\n});"
			}`,
			ExpectedStatus: http.StatusOK,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Save(mock.Anything, "user123", "conv456").Return(nil).Once()
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
		},
		{
			TestName: "Save service fails",
			RequestBody: `{
				"userId": "user789",
				"conversationId": "conv789",
				"code": "test code"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Save(mock.Anything, "user789", "conv789").Return(errors.New("database error")).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidatePrompt(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)

			test.SetupMock(mockLocalStorageServ)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/saveLocal", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ)
			if err != nil {
				t.Errorf("build failed")
			}

			controller.HandleSaveLocalRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}

// nolint:dupl
func TestHandleDeleteLocalRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		SetupMock      func(*mocks.MockTestcaseLocalStorageService)
	}{
		{
			TestName: "Valid delete request",
			RequestBody: `{
				"testcaseId": "test123",
				"userId": "user123",
				"conversationId": "conv456"
			}`,
			ExpectedStatus: http.StatusOK,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Delete("test123", "user123", "conv456").Return(nil).Once()
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":json}`,
			ExpectedStatus: http.StatusBadRequest,
			SetupMock:      func(m *mocks.MockTestcaseLocalStorageService) {},
		},
		{
			TestName: "Delete service fails",
			RequestBody: `{
				"testcaseId": "test789",
				"userId": "user789",
				"conversationId": "conv789"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			SetupMock: func(m *mocks.MockTestcaseLocalStorageService) {
				m.EXPECT().Delete("test789", "user789", "conv789").Return(errors.New("database error")).Once()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidatePrompt(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)

			test.SetupMock(mockLocalStorageServ)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/deleteLocal", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ)
			if err != nil {
				t.Errorf("build failed")
			}

			controller.HandleDeleteLocalRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}
