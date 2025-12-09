package handler

import (
	"bytes"
	"context"
	"encoding/json"
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
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", UserId: "2"}, nil}},
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", UserId: "2"}, nil}},
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", UserId: "2"}, nil}},
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", UserId: "2"}, nil}},
				{Function: "ValidatePrompt", ExpectedResponse: []any{true, "", nil}},
				{Function: "GeneratePrompt", ExpectedResponse: []any{"", errors.New("err")}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)

			for _, mc := range test.MockSetup {
				switch mc.Function {
				case "ValidatePrompt":
					mockValServ.
						On("ValidatePrompt", mock.Anything, mock.Anything, mock.Anything).
						Return(mc.ExpectedResponse...)
				case "GeneratePrompt":
					mockGenServ.
						On("GeneratePrompt", mock.Anything, mock.Anything, mock.Anything).
						Return(mc.ExpectedResponse...)
				case "LoadChat":
					mockChatManager.
						On("LoadChat", mock.Anything, mock.Anything).
						Return(mc.ExpectedResponse...)
				case "SaveChat":
					mockChatManager.
						On("SaveChat", mock.Anything, mock.Anything).
						Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockChatManager)

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
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)

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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockChatManager)

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

func TestGetUserChats_ValidationCases(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name           string
		requestID      string
		limit          string
		expectedStatus int
	}{
		{"error - No ID", "", "", http.StatusNotFound},
		{"error - invalid ID format", "132", "", http.StatusBadRequest},
		{"error - limit is not a number", "auth0|123", "hallo", http.StatusBadRequest},
		{"error - limit is a negative number", "auth0|123", "-131", http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)

			router := gin.New()

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockChatManager)

			router.GET("/api/v1/users/:userId/chats", controller.HandleGetUserChats)

			var endpoint string
			if tc.requestID == "" {
				endpoint = "/api/v1/users/"
			} else {
				endpoint = "/api/v1/users/" + tc.requestID + "/chats"
				if tc.limit != "" {
					endpoint += "?limit=" + tc.limit
				}
			}

			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestGetUserChats_LoadCases(t *testing.T) {
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
			"error - no chat history for given user found",
			"auth0|123",
			"123",
			[]any{nil, errors.New("no history found")},
			http.StatusInternalServerError,
		},
		{
			"success",
			"auth0|123",
			"5",
			[]any{[]*entity.ChatSummary{
				{ChatId: "1", UpdatedAt: time.Now()},
				{ChatId: "2", UpdatedAt: time.Now().Add(-10 * time.Hour)},
			}, nil},
			http.StatusOK,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
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

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockChatManager)
			router.GET("/api/v1/users/:userId/chats", controller.HandleGetUserChats)

			endpoint := "/api/v1/users/" + tc.requestID + "/chats"
			if tc.limit != "" {
				endpoint += "?limit=" + tc.limit
			}

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

func newTestControllerWithChatMock(t *testing.T, chat *entity.Chat, err error) *AutotesterController {
	t.Helper()

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockChatManager := mocks.NewMockChatManager(t)

	mockChatStorageServ.
		On("LoadChat", mock.Anything, mock.Anything, mock.Anything).
		Return(chat, err)

	controller, buildErr := NewAutotesterController(
		logger,
		cfg,
		mockValServ,
		mockGenServ,
		mockLocalStorageServ,
		mockDockerServ,
		mockChatStorageServ,
		mockRemoteStorageServ,
		mockChatManager,
	)
	if buildErr != nil {
		t.Fatalf("failed to build controller: %v", buildErr)
	}

	return controller
}

func TestGetChatById_MissingParams_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockChatManager := mocks.NewMockChatManager(t)

	controller, err := NewAutotesterController(
		logger,
		cfg,
		mockValServ,
		mockGenServ,
		mockLocalStorageServ,
		mockDockerServ,
		mockChatStorageServ,
		mockRemoteStorageServ,
		mockChatManager,
	)
	if err != nil {
		t.Fatalf("failed to build controller: %v", err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/chats/someChat", nil)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req

	controller.GetChatById(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetChatById_ChatNotFound_ReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUserID := "auth0|user-42"
	validChatID := "550e8400-e29b-41d4-a716-446655440000"

	controller := newTestControllerWithChatMock(t, nil, sharedErrors.ErrChatNotFound)

	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/users/"+validUserID+"/chats/"+validChatID,
		nil,
	)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "userId", Value: validUserID},
		{Key: "chatId", Value: validChatID},
	}

	controller.GetChatById(ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetChatById_Success_ReturnsChat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validUserID := "auth0|user-42"
	validChatID := "550e8400-e29b-41d4-a716-446655440000"

	expectedChat := &entity.Chat{
		Id:     validChatID,
		UserId: validUserID,
	}

	controller := newTestControllerWithChatMock(t, expectedChat, nil)

	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/users/"+validUserID+"/chats/"+validChatID,
		nil,
	)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "userId", Value: validUserID},
		{Key: "chatId", Value: validChatID},
	}

	controller.GetChatById(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp entity.Chat
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Id != expectedChat.Id || resp.UserId != expectedChat.UserId {
		t.Fatalf("unexpected chat in response: %+v", resp)
	}
}
