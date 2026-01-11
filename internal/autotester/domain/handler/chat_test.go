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

	"go.opentelemetry.io/otel"

	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"

	"github.com/gin-gonic/gin"
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
	tracer := otel.Tracer("test")

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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", Author: "2", LastModifiedBy: "2"}, nil}},
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", Author: "2", LastModifiedBy: "2"}, nil}},
				{Function: "GeneratePrompt", ExpectedResponse: []any{"", errors.New("err")}},
			},
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

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
			mockGroups := mocks.NewMockGroupStorage(t)

			for _, mc := range test.MockSetup {
				switch mc.Function {
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
						On("SaveChat", mock.Anything, mock.Anything, mock.Anything).
						Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			controller, _ := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockChatManager,
				mockGroups,
				tracer,
				mockMetricsServ,
			)

			controller.HandleChatRequest(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}

// nolint:funlen
func TestHandleChatRequestValidity(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	tests := []struct {
		TestName             string
		RequestBody          string
		ExpectedStatus       int
		MockResponseValidate []any
		MockResponseLoad     []any
		MockResponseSave     []any
	}{
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid":it ad json}`,
			ExpectedStatus: http.StatusBadRequest,
		},
		{
			TestName: "valid request, valid prompt",
			RequestBody: `{
				"message": {
					"body": "this is a valid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus:       http.StatusOK,
			MockResponseLoad:     []any{&entity.Chat{}, nil},
			MockResponseValidate: []any{true, "", nil},
			MockResponseSave:     []any{nil},
		},
		{
			TestName: "valid request, invalid prompt",
			RequestBody: `{
				"message": {
					"body":"this is a invalid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus:       http.StatusOK,
			MockResponseLoad:     []any{&entity.Chat{}, nil},
			MockResponseValidate: []any{false, "Invalid prompt because...", nil},
			MockResponseSave:     []any{nil},
		},
		{
			TestName: "valid request, errors when validating",
			RequestBody: `{
				"message": {
					"body":"validating err",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus:       http.StatusInternalServerError,
			MockResponseLoad:     []any{&entity.Chat{}, nil},
			MockResponseValidate: []any{true, "", errors.New("err")},
		},
		{
			TestName: "assert fails",
			RequestBody: `{
				"message": {
					"body": "this is a valid prompt",
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
					"body": "this is a valid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus:   http.StatusInternalServerError,
			MockResponseLoad: []any{nil, errors.New("err")},
		},
		{
			TestName: "SaveChat error",
			RequestBody: `{
				"message": {
					"body": "this is a valid prompt",
					"role":"user"
				},
				"userId":"2",
				"conversationId":"2"
			}`,
			ExpectedStatus:       http.StatusOK,
			MockResponseLoad:     []any{&entity.Chat{}, nil},
			MockResponseValidate: []any{true, "", nil},
			MockResponseSave:     []any{errors.New("err")},
		},
	}
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockValServ := mocks.NewMockValidator(t)
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)
			mockGroups := mocks.NewMockGroupStorage(t)

			if test.MockResponseLoad != nil {
				mockChatManager.On("LoadChat", mock.Anything, mock.Anything).Return(test.MockResponseLoad...)
			}
			if test.MockResponseSave != nil {
				mockChatManager.On("SaveChat", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseSave...)
			}
			if test.MockResponseValidate != nil {
				mockValServ.On("ValidatePrompt", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseValidate...)
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat/validity", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ,
				mockChatStorageServ, mockRemoteStorageServ, mockChatManager, mockGroups, tracer, mockMetricsServ)
			controller.HandleChatRequestValidity(ctx)
			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
		})
	}
}

func TestHandleUserInfoRequest(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
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
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockGroups := mocks.NewMockGroupStorage(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

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
				mockGroups,
				tracer,
				mockMetricsServ,
			)

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

//nolint:funlen
func TestGetUserChats(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	tests := []struct {
		name             string
		limit            string
		mockResponseLoad []any
		expectedStatus   int
	}{
		{
			name:           "error - limit is not a number",
			limit:          "hallo",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit is a negative number",
			limit:          "-131",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:             "error - no chat history for given user found",
			limit:            "123",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "error - empty limit but no history found",
			limit:            "",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:  "success",
			limit: "5",
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
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatManager := mocks.NewMockChatManager(t)
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockGroups := mocks.NewMockGroupStorage(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadSummaries", mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
			}
			gin.SetMode(gin.TestMode)
			router := gin.New()

			controller, _ := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockChatManager,
				mockGroups,
				tracer,
				mockMetricsServ,
			)
			router.GET("/api/v1/chats/", controller.HandleGetChats)

			endpoint := "/api/v1/chats/?limit=" + tc.limit
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
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
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockGroups := mocks.NewMockGroupStorage(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	tracer := otel.Tracer("test")

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
		mockGroups,
		tracer,
		mockMetricsServ,
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
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockGroups := mocks.NewMockGroupStorage(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	tracer := otel.Tracer("test")

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
		mockGroups,
		tracer,
		mockMetricsServ,
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
		Id:             validChatID,
		Author:         validUserID,
		LastModifiedBy: validChatID,
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

	if resp.Id != expectedChat.Id || resp.Author != expectedChat.Author || resp.LastModifiedBy != expectedChat.LastModifiedBy {
		t.Fatalf("unexpected chat in response: %+v", resp)
	}
}
