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
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)

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

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ)

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

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ)

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

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			gin.SetMode(gin.TestMode)
			router := gin.New()

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ)
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadUserChats", mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
			}

			gin.SetMode(gin.TestMode)
			router := gin.New()

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ, mockChatStorageServ, mockRemoteStorageServ)
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

	mockChatStorageServ.EXPECT().
		LoadChat(mock.Anything, mock.Anything, mock.Anything).
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

	controller, err := NewAutotesterController(
		logger,
		cfg,
		mockValServ,
		mockGenServ,
		mockLocalStorageServ,
		mockDockerServ,
		mockChatStorageServ,
		mockRemoteStorageServ,
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

	// gültige IDs, die die neue Validierung bestehen
	validUserID := "auth0|user-42"
	validChatID := "550e8400-e29b-41d4-a716-446655440000" // irgendeine gültige UUID

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
