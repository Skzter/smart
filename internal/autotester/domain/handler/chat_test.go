package handler

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service/mocks"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

// nolint:funlen
func TestHandleChatRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	validPrompt := "this is a valid prompt"
	invalidPrompt := "this is a invalid prompt"

	type MockSetup struct {
		Function         string
		UserPrompt       string
		ExpectedResponse string
		ExpectedBool     bool
		ResponseError    error
	}

	tests := []struct {
		TestName       string
		RequestBody    string
		ExpectedStatus int
		MockSetup      []MockSetup
	}{
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":it ad json}`,
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName: "valid request",
			RequestBody: `{
				"message": {
					"body":"this is a valid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "ValidatePrompt", UserPrompt: validPrompt, ExpectedBool: true},
				{Function: "GeneratePrompt", UserPrompt: validPrompt, ExpectedResponse: "some code"},
			},
		},
		{
			TestName: "sessionId is missing and controller must generate one",
			RequestBody: `{
				"message": {
					"body":"this is a valid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":""
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "ValidatePrompt", UserPrompt: validPrompt, ExpectedBool: true},
				{Function: "GeneratePrompt", UserPrompt: validPrompt, ExpectedResponse: "some code"},
			},
		},
		{
			TestName: "invalid request => invalid prompt",
			RequestBody: `{
				"message": {
					"body":"this is a invalid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "ValidatePrompt", UserPrompt: invalidPrompt, ExpectedBool: false, ExpectedResponse: "versuch doch mal das"},
			},
		},
		{
			TestName: "valid request, validate will return false json",
			RequestBody: `{
				"message": {
					"body":"json gibts nicht",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: []MockSetup{
				{Function: "ValidatePrompt", UserPrompt: "json gibts nicht", ExpectedBool: false, ResponseError: sharedErrors.ErrValidation},
			},
		},
		{
			TestName: "valid request, errors when generating",
			RequestBody: `{
				"message": {
					"body":"generating err",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: []MockSetup{
				{Function: "ValidatePrompt", UserPrompt: "generating err", ExpectedBool: true},
				{Function: "GeneratePrompt", UserPrompt: "generating err", ResponseError: sharedErrors.ErrGeneration},
			},
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			for _, mc := range test.MockSetup {
				switch mc.Function {
				case "ValidatePrompt":
					mockValServ.EXPECT().
						ValidatePrompt(mock.Anything, mc.UserPrompt).
						Return(mc.ExpectedBool, mc.ExpectedResponse, mc.ResponseError)
				case "GeneratePrompt":
					mockGenServ.EXPECT().
						GeneratePrompt(mock.Anything, mc.UserPrompt).
						Return(mc.ExpectedResponse, mc.ResponseError)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ)

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
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)

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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ)

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

func TestGetUserChats(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tests := []struct {
		name             string
		requestID        string
		limit            string
		mockResponseLoad []any
		expectedStatus   int
	}{
		{
			name:           "error - No ID",
			requestID:      "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "error - invalid ID format",
			requestID:      "132",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit is not a number",
			requestID:      "auth0|123",
			limit:          "hallo",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit is a negative number",
			requestID:      "auth0|123",
			limit:          "-131",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:             "error - no chat history for given user found",
			requestID:        "auth0|123",
			limit:            "123",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "error - empty limit but no history found",
			requestID:        "auth0|123",
			limit:            "",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:      "success",
			requestID: "auth0|123",
			limit:     "5",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "1",
					UpdatedAt: time.Now(),
				},
				{
					ChatId:    "2",
					UpdatedAt: time.Now().Add(-10 * time.Hour),
				},
			}, nil},
			expectedStatus: http.StatusOK,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidatePrompt(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadUserChats", mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
			}
			gin.SetMode(gin.TestMode)
			router := gin.New()

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ)
			router.GET("/api/v1/chats/:UserID", controller.HandleGetUserChats)

			endpoint := "/api/v1/chats/" + tc.requestID + "?limit=" + tc.limit
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
