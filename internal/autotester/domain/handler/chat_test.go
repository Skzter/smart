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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

const validUserId = "auth0|user-42"
const validChatId = "550e8400-e29b-41d4-a716-446655440000"

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
			TestName: "Valid request",
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
				{Function: "LoadChat", ExpectedResponse: []any{&entity.Chat{Id: "2", UserId: "2"}, nil}},
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
						On("SaveChat", mock.Anything, mock.Anything).
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
func TestHandleChatRequestValIdity(t *testing.T) {
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
				"userid":"",
				"conversationid":"2"
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
				"userid":"2",
				"conversationid":"2"
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
				"userid":"2",
				"conversationid":"2"
			}`,
			ExpectedStatus:       http.StatusOK,
			MockResponseLoad:     []any{&entity.Chat{}, nil},
			MockResponseValidate: []any{true, "", nil},
			MockResponseSave:     []any{errors.New("err")},
		},
	}
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
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

			if test.MockResponseLoad != nil {
				mockChatManager.On("LoadChat", mock.Anything, mock.Anything).Return(test.MockResponseLoad...)
			}
			if test.MockResponseSave != nil {
				mockChatManager.On("SaveChat", mock.Anything, mock.Anything).Return(test.MockResponseSave...)
			}
			if test.MockResponseValidate != nil {
				mockValServ.On("ValidatePrompt", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseValidate...)
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat/validate", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ,
				mockChatStorageServ, mockRemoteStorageServ, mockChatManager, tracer, mockMetricsServ)
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
				"userid": "9177b856-46a0-11f0-9fe2-0242ac120002",
				"allConversations": [
				  {
					"Conversationid": "string",
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
			TestName:        "InvalId UserRequestBody",
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
		requestId        string
		limit            string
		mockResponseLoad []any
		expectedStatus   int
	}{
		{
			name:           "error - No Id",
			requestId:      "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "error - invalId Id format",
			requestId:      "132",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit is not a number",
			requestId:      "auth0|123",
			limit:          "hallo",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error - limit is a negative number",
			requestId:      "auth0|123",
			limit:          "-131",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:             "error - no chat history for given user found",
			requestId:        "auth0|123",
			limit:            "123",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "error - empty limit but no history found",
			requestId:        "auth0|123",
			limit:            "",
			mockResponseLoad: []any{nil, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:      "success",
			requestId: "auth0|123",
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
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatManager := mocks.NewMockChatManager(t)
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadUserChats", mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
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
				tracer,
				mockMetricsServ,
			)
			router.GET("/api/v1/chats/:userId", controller.HandleGetUserChats)

			endpoint := "/api/v1/chats/" + tc.requestId + "?limit=" + tc.limit
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestIsValId(t *testing.T) {
	tests := []struct {
		name      string
		Id        string
		expectErr bool
	}{
		{
			name:      "error - Id doesnt have separator",
			Id:        "auth01234",
			expectErr: true,
		},
		{
			name:      "error - first token isnt 'auth0'",
			Id:        "hallo|123",
			expectErr: true,
		},
		{
			name:      "error - second token is empty string",
			Id:        "hallo|123",
			expectErr: true,
		},
		{
			name:      "success",
			Id:        "auth0|123",
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			valid := isValid(tc.Id)
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
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)

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

	controller := newTestControllerWithChatMock(t, nil, sharedErrors.ErrChatNotFound)

	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/users/"+validUserId+"/chats/"+validChatId,
		nil,
	)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "userId", Value: validUserId},
		{Key: "chatId", Value: validChatId},
	}

	controller.GetChatById(ctx)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetChatById_Success_ReturnsChat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedChat := &entity.Chat{
		Id:     validChatId,
		UserId: validUserId,
	}

	controller := newTestControllerWithChatMock(t, expectedChat, nil)

	req, _ := http.NewRequest(
		http.MethodGet,
		"/api/v1/users/"+validUserId+"/chats/"+validChatId,
		nil,
	)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = gin.Params{
		{Key: "userId", Value: validUserId},
		{Key: "chatId", Value: validChatId},
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

func TestHandleUpdateChatTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	tests := []struct {
		TestName          string
		UserId            string
		ChatId            string
		RequestBody       string
		MockResponseLoad  []*entity.Chat
		MockErrorLoad     []error
		MockErrorSave     []error
		ExpectedStatus    int
		ExpectedErrorText string
	}{
		{
			TestName:    "Successful update",
			UserId:      validUserId,
			ChatId:      validChatId,
			RequestBody: `{"title":"New Chat Title"}`,
			MockResponseLoad: []*entity.Chat{
				{Id: validChatId, UserId: validUserId, Title: "Old Title"},
			},
			MockErrorLoad:  []error{nil},
			MockErrorSave:  []error{nil},
			ExpectedStatus: http.StatusOK,
		},
		{
			TestName:          "Invalid user Id",
			UserId:            "!!invalid!!",
			ChatId:            validChatId,
			RequestBody:       `{"title":"New Chat Title"}`,
			ExpectedStatus:    http.StatusBadRequest,
			ExpectedErrorText: "invalid userId format",
		},
		{
			TestName:    "LoadChat fails",
			UserId:      validUserId,
			ChatId:      validChatId,
			RequestBody: `{"title":"New Chat Title"}`,
			MockResponseLoad: []*entity.Chat{
				{},
			},
			MockErrorLoad:     []error{errors.New("db error")},
			ExpectedStatus:    http.StatusInternalServerError,
			ExpectedErrorText: "could not load chat",
		},
	}
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			mockGenServ := mocks.NewMockGeneratePrompt(t)
			mockValServ := mocks.NewMockValidator(t)
			mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
			mockDockerServ := mocks.NewMockDocker(t)
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
			mockChatManager := mocks.NewMockChatManager(t)

			if test.MockResponseLoad != nil {
				mockChatStorageServ.On("LoadChat", mock.Anything, test.UserId, test.ChatId).Return(test.MockResponseLoad[0], test.MockErrorLoad[0])
			}
			if test.MockErrorSave != nil {
				mockChatManager.On("SaveChat", mock.Anything, mock.Anything).Return(test.MockErrorSave[0]).Maybe()
			}

			req, _ := http.NewRequest(http.MethodPut, "/users/"+test.UserId+"/chats/"+test.ChatId, bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Params = []gin.Param{
				{Key: "userId", Value: test.UserId},
				{Key: "chatId", Value: test.ChatId},
			}

			controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ, mockDockerServ,
				mockChatStorageServ, mockRemoteStorageServ, mockChatManager, tracer, mockMetricsServ)
			controller.HandleUpdateChatTitle(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}
			if test.ExpectedErrorText != "" {
				var resp entity.ErrorMessage
				_ = json.Unmarshal(rec.Body.Bytes(), &resp)
				if resp.Error != test.ExpectedErrorText {
					t.Errorf("Expected error '%s', got '%s'", test.ExpectedErrorText, resp.Error)
				}
			}
		})
	}
}

func TestHandleUpdateChatTitle_TitleTooLong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	validUserId := "auth0|user-42"
	validChatId := uuid.NewString()

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatStorageServ := mocks.NewMockChatStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockChatManager := mocks.NewMockChatManager(t)

	mockChatStorageServ.On("LoadChat", mock.Anything, validUserId, validChatId).
		Return(&entity.Chat{
			Id:     validChatId,
			UserId: validUserId,
			Title:  "Old Title",
		}, nil)

	mockChatManager.On("SaveChat", mock.Anything, mock.Anything).Return(nil).Maybe()

	requestBody := `{"title":"This title is way too long to be accepted by the system"}`
	req, _ := http.NewRequest(http.MethodPut, "/users/"+validUserId+"/chats/"+validChatId, bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	ctx.Params = []gin.Param{
		{Key: "userId", Value: validUserId},
		{Key: "chatId", Value: validChatId},
	}

	controller, _ := NewAutotesterController(logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ,
		mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockChatManager, tracer, mockMetricsServ)

	controller.HandleUpdateChatTitle(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp entity.ErrorMessage
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	expectedError := "Title must be 1–30 characters"
	if resp.Error != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, resp.Error)
	}
}
