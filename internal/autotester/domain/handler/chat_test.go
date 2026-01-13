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

const validUserID = "auth0|user-42"
const validChatID = "550e8400-e29b-41d4-a716-446655440000"

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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

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

			mockMediaServ := mocks.NewMockMediaStorageService(t)
			controller, _ := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
				"chatId":"2"
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
			mockAuth := mocks.NewMockAuth(t)
			mockGroupManager := mocks.NewMockGroupManager(t)

			if test.MockResponseLoad != nil {
				mockChatManager.On("LoadChat", mock.Anything, mock.Anything).Return(test.MockResponseLoad...)
			}
			if test.MockResponseSave != nil {
				mockChatManager.On("SaveChat", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseSave...)
			}
			if test.MockResponseValidate != nil {
				mockValServ.On("ValidatePrompt", mock.Anything, mock.Anything, mock.Anything).Return(test.MockResponseValidate...)
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chat/validate", bytes.NewBufferString(test.RequestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			mockMediaServ := mocks.NewMockMediaStorageService(t)
			controller, _ := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
			)
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
				"allChats": [
				  {
					"ChatId": "string",
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
	mockAuth := mocks.NewMockAuth(t)
	mockGroupManager := mocks.NewMockGroupManager(t)

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

			mockMediaServ := mocks.NewMockMediaStorageService(t)
			controller, err := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
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
	t.Skip()
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")
	tests := []struct {
		name             string
		queryParams      string
		mockResponseLoad []any
		expectedStatus   int
		checkPageSize    bool
	}{
		{
			name:             "error - no chat history for given user found",
			queryParams:      "",
			mockResponseLoad: []any{nil, false, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "error - empty limit but no history found",
			queryParams:      "",
			mockResponseLoad: []any{nil, false, errors.New("no history found")},
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:        "success",
			queryParams: "",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "1",
					UpdatedAt: time.Now(),
				},
				{
					ChatId:    "2",
					UpdatedAt: time.Now().Add(-10 * time.Hour),
				},
			}, false, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
		{
			name:        "success with hasMore true",
			queryParams: "",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "1",
					UpdatedAt: time.Now(),
				},
				{
					ChatId:    "2",
					UpdatedAt: time.Now().Add(-10 * time.Hour),
				},
			}, true, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
		{
			name:             "success with empty result",
			queryParams:      "",
			mockResponseLoad: []any{[]*entity.ChatSummary{}, false, nil},
			expectedStatus:   http.StatusOK,
			checkPageSize:    true,
		},
		{
			name:        "success with page 0",
			queryParams: "?page=0",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "1",
					UpdatedAt: time.Now(),
				},
			}, false, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
		{
			name:        "success with page 1",
			queryParams: "?page=1",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "3",
					UpdatedAt: time.Now(),
				},
			}, true, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
		{
			name:           "error - invalid page parameter (negative)",
			queryParams:    "?page=-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "success with valid groups",
			queryParams: "?groups=group1&groups=group2",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "1",
					UpdatedAt: time.Now(),
				},
			}, false, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
		{
			name:           "error - invalid groups parameter (empty string)",
			queryParams:    "?groups=",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "success with page and groups combined",
			queryParams: "?page=2&groups=group1&groups=group2",
			mockResponseLoad: []any{[]*entity.ChatSummary{
				{
					ChatId:    "5",
					UpdatedAt: time.Now(),
				},
			}, true, nil},
			expectedStatus: http.StatusOK,
			checkPageSize:  true,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidator(t)
	mockLocalStorageServ := mocks.NewMockTestcaseLocalStorageService(t)
	mockRemoteStorageServ := mocks.NewMockTestcaseStorageService(t)
	mockDockerServ := mocks.NewMockDocker(t)
	mockChatManager := mocks.NewMockChatManager(t)
	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockAuth := mocks.NewMockAuth(t)
	mockGroupManager := mocks.NewMockGroupManager(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockChatStorageServ := mocks.NewMockChatStorageService(t)
			if tc.mockResponseLoad != nil {
				mockChatStorageServ.On("LoadSummaries", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mockResponseLoad...)
			}
			gin.SetMode(gin.TestMode)
			router := gin.New()

			mockMediaServ := mocks.NewMockMediaStorageService(t)
			controller, _ := NewAutotesterController(
				logger,
				cfg,
				mockValServ,
				mockGenServ,
				mockLocalStorageServ,
				mockDockerServ,
				mockChatStorageServ,
				mockRemoteStorageServ,
				mockMediaServ,
				mockChatManager,
				mockGroupManager,
				tracer,
				mockMetricsServ,
				mockAuth,
			)
			router.GET("/api/v1/chats/", controller.HandleGetChats)

			endpoint := "/api/v1/chats/" + tc.queryParams
			req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			// Check that PageSize is included in successful responses
			if tc.checkPageSize && rec.Code == http.StatusOK {
				var response entity.ChatSummarys
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.PageSize != cfg.PageSize {
					t.Errorf("Expected PageSize %d, got %d", cfg.PageSize, response.PageSize)
				}
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
	mockAuth := mocks.NewMockAuth(t)
	mockGroupManager := mocks.NewMockGroupManager(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	tracer := otel.Tracer("test")

	mockChatStorageServ.
		On("LoadChat", mock.Anything, mock.Anything, mock.Anything).
		Return(chat, err)

	mockMediaServ := mocks.NewMockMediaStorageService(t)
	controller, buildErr := NewAutotesterController(
		logger,
		cfg,
		mockValServ,
		mockGenServ,
		mockLocalStorageServ,
		mockDockerServ,
		mockChatStorageServ,
		mockRemoteStorageServ,
		mockMediaServ,
		mockChatManager,
		mockGroupManager,
		tracer,
		mockMetricsServ,
		mockAuth,
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
	mockAuth := mocks.NewMockAuth(t)
	mockGroupManager := mocks.NewMockGroupManager(t)

	// Setup metrics mock to accept any calls
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordStatusCode", mock.Anything).Return().Maybe()

	tracer := otel.Tracer("test")

	mockMediaServ := mocks.NewMockMediaStorageService(t)
	controller, err := NewAutotesterController(
		logger,
		cfg,
		mockValServ,
		mockGenServ,
		mockLocalStorageServ,
		mockDockerServ,
		mockChatStorageServ,
		mockRemoteStorageServ,
		mockMediaServ,
		mockChatManager,
		mockGroupManager,
		tracer,
		mockMetricsServ,
		mockAuth,
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

// nolint:funlen
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
		MockLoadReturn    []any
		MockSaveReturn    []any
		ExpectedStatus    int
		ExpectedErrorText string
	}{
		{
			TestName:    "Successful update",
			UserId:      validUserID,
			ChatId:      validChatID,
			RequestBody: `{"title":"New Chat Title"}`,

			MockLoadReturn: []any{
				&entity.Chat{Id: validChatID, Author: validUserID, Title: "Old Title"},
				nil,
			},
			MockSaveReturn: []any{nil},
			ExpectedStatus: http.StatusOK,
		},
		{
			TestName:    "LoadChat fails",
			UserId:      validUserID,
			ChatId:      validChatID,
			RequestBody: `{"title":"New Chat Title"}`,
			MockLoadReturn: []any{
				nil,
				errors.New("db error"),
			},
			MockSaveReturn:    nil,
			ExpectedStatus:    http.StatusInternalServerError,
			ExpectedErrorText: "could not load chat",
		},
		{
			TestName:    "SaveChat fails",
			UserId:      validUserID,
			ChatId:      validChatID,
			RequestBody: `{"title":"New Chat Title"}`,
			MockLoadReturn: []any{
				&entity.Chat{Id: validChatID, Author: validUserID, Title: "Old Title"},
				nil,
			},
			MockSaveReturn:    []any{errors.New("db error")},
			ExpectedStatus:    http.StatusInternalServerError,
			ExpectedErrorText: "could not save chat",
		},
		{
			TestName:    "Title too long",
			UserId:      validUserID,
			ChatId:      validChatID,
			RequestBody: `{"title":"1234567890123456789012345678901"}`,
			MockLoadReturn: []any{
				&entity.Chat{Id: validChatID, Author: validUserID, Title: "Old Title"},
				nil,
			},
			MockSaveReturn:    []any{nil},
			ExpectedStatus:    http.StatusBadRequest,
			ExpectedErrorText: "Title must be 1–30 characters",
		},
		{
			TestName:    "Empty title",
			UserId:      validUserID,
			ChatId:      validChatID,
			RequestBody: `{"title":" "}`,
			MockLoadReturn: []any{
				&entity.Chat{Id: validChatID, Author: validUserID, Title: "Old Title"},
				nil,
			},
			MockSaveReturn:    []any{nil},
			ExpectedStatus:    http.StatusBadRequest,
			ExpectedErrorText: "Title must be 1–30 characters",
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
			mockGroupManager := mocks.NewMockGroupManager(t)
			mockAuth := mocks.NewMockAuth(t)
			mockMediaService := mocks.NewMockMediaStorageService(t)

			if test.MockLoadReturn != nil {
				mockChatStorageServ.
					On("LoadChat", mock.Anything, mock.Anything).
					Return(test.MockLoadReturn...)
			}

			if test.MockSaveReturn != nil {
				mockChatManager.
					On("SaveChat", mock.Anything, mock.Anything, mock.Anything).
					Return(test.MockSaveReturn...).
					Maybe()
			}

			router := gin.New()

			controller, _ := NewAutotesterController(
				logger, cfg, mockValServ, mockGenServ, mockLocalStorageServ,
				mockDockerServ, mockChatStorageServ, mockRemoteStorageServ, mockMediaService,
				mockChatManager, mockGroupManager, tracer, mockMetricsServ, mockAuth,
			)

			router.PATCH("/users/:userId/chats/:chatId", controller.HandleUpdateChatTitle)

			req, _ := http.NewRequest(
				http.MethodPatch,
				"/users/"+test.UserId+"/chats/"+test.ChatId,
				bytes.NewBufferString(test.RequestBody),
			)
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

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

			if test.ExpectedStatus == http.StatusOK {
				var resp map[string]string
				_ = json.Unmarshal(rec.Body.Bytes(), &resp)

				if resp["chatId"] != validChatID {
					t.Errorf("Expected chatId '%s', got '%s'", validChatID, resp["chatId"])
				}
				if resp["title"] != "New Chat Title" {
					t.Errorf("Expected title 'New Chat Title', got '%s'", resp["title"])
				}
			}
		})
	}
}
