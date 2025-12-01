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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

// nolint:funlen
func TestHandleChatRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	type MockSetup struct {
		Function         string
		ExpectedResponse []any
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
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{true, "", nil}},
				{Function: "GeneratePrompt", ExpectedResponse: []any{"some code", nil}},
				{Function: "SaveChat", ExpectedResponse: []any{nil}},
			},
		},
		{
			TestName: "invalid userId",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "loadChat error",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{nil, errors.New("err")}},
			},
		},
		{
			TestName: "validate error",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{false, "", errors.New("err")}},
			},
		},
		{
			TestName: "invalid message",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{false, "invalid message", nil}},
				{Function: "SaveChat", ExpectedResponse: []any{nil}},
			},
		},
		{
			TestName: "saveChat error",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{true, "", nil}},
				{Function: "GeneratePrompt", ExpectedResponse: []any{"some code", nil}},
				{Function: "SaveChat", ExpectedResponse: []any{errors.New("err")}},
			},
		},
		{
			TestName: "generate error",
			RequestBody: `{
				"message": {
					"body":"prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
			MockSetup: []MockSetup{
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{true, "", nil}},
				{Function: "GeneratePrompt", ExpectedResponse: []any{"", errors.New("err")}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)
			for _, mc := range test.MockSetup {
				switch mc.Function {
				case "ValidatePrompt":
					mockValServ.On(mc.Function, mock.Anything, mock.Anything, mock.Anything).Return(mc.ExpectedResponse...)
				case "GeneratePrompt":
					mockGenServ.On(mc.Function, mock.Anything, mock.Anything, mock.Anything).Return(mc.ExpectedResponse...)
				case "LoadChat":
					mockChatManager.On(mc.Function, mock.Anything, mock.Anything).Return(mc.ExpectedResponse...)
				case "SaveChat":
					mockChatManager.On(mc.Function, mock.Anything, mock.Anything).Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockChatManager)

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
	mockChatManager := mocks.NewMockChatManager(t)
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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockChatManager)

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
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatManager := mocks.NewMockChatManager(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadUserChats", mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
			}
			gin.SetMode(gin.TestMode)
			router := gin.New()

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockChatManager)
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

func TestIsValid(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		expectErr bool
	}{
		{
			name:      "error - id doesnt have separator",
			id:        "auth01234",
			expectErr: true,
		},
		{
			name:      "error - first token isnt 'auth0'",
			id:        "hallo|123",
			expectErr: true,
		},
		{
			name:      "error - second token is empty string",
			id:        "hallo|123",
			expectErr: true,
		},
		{
			name:      "success",
			id:        "auth0|123",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := isValid(tc.id)
			if tc.expectErr {
				assert.False(t, valid)
			} else {
				assert.True(t, valid)
			}
		})
	}
}
