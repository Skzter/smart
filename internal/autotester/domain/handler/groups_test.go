package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	sharedMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
)

func TestHandleGetGroups(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	type MockSetup struct {
		Function         string
		ExpectedResponse []any
	}

	tests := []struct {
		TestName       string
		ExpectedStatus int
		MockSetup      []MockSetup
	}{
		{
			TestName:       "Successfully retrieve groups",
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "List", ExpectedResponse: []any{[]*entity.Group{{Id: "group1", Name: "Test Group"}}, nil}},
			},
		},
		{
			TestName:       "Group manager error",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: []MockSetup{
				{Function: "List", ExpectedResponse: []any{nil, errors.New("storage error")}},
			},
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()

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
			mockGroupManager := mocks.NewMockGroupManager(t)

			for _, mc := range test.MockSetup {
				if mc.Function == "List" {
					mockGroupManager.
						On("List", mock.Anything).
						Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodGet, "/api/v1/groups", nil)
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
				mockGroupManager,
				tracer,
				mockMetricsServ,
			)

			controller.HandleGetGroups(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			mockGroupManager.AssertExpectations(t)
		})
	}
}

//nolint:funlen
func TestHandleCreateGroup(t *testing.T) {
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
			TestName: "Successfully create group",
			RequestBody: `{
				"userId": "user123",
				"groupName": "Test Group",
				"description": "A test group"
			}`,
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "Create", ExpectedResponse: []any{"group-uuid-123", nil}},
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid": json}`,
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName: "Group manager create error",
			RequestBody: `{
				"userId": "user123",
				"groupName": "Test Group",
				"description": "A test group"
			}`,
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: []MockSetup{
				{Function: "Create", ExpectedResponse: []any{"", errors.New("creation error")}},
			},
		},
		{
			TestName: "Missing required fields",
			RequestBody: `{
				"description": "A test group"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()

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
			mockGroupManager := mocks.NewMockGroupManager(t)

			for _, mc := range test.MockSetup {
				if mc.Function == "Create" {
					mockGroupManager.
						On("Create", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewBufferString(test.RequestBody))
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
				mockGroupManager,
				tracer,
				mockMetricsServ,
			)

			controller.HandleCreateGroup(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			// Verify response content for success cases
			if test.ExpectedStatus == http.StatusOK && test.MockSetup != nil {
				var response entity.CreateGroupResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}
				if response.GroupId == "" {
					t.Errorf("Expected non-empty GroupId in response")
				}
			}

			if test.MockSetup != nil {
				mockGroupManager.AssertExpectations(t)
			}
		})
	}
}

//nolint:funlen
func TestHandleAssignChatToGroups(t *testing.T) {
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
		ChatId         string
		ExpectedStatus int
		MockSetup      []MockSetup
	}{
		{
			TestName: "Successfully assign chat to single group",
			RequestBody: `{
				"groupIds": ["550e8400-e29b-41d4-a716-446655440000"]
			}`,
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "AddChatToGroup", ExpectedResponse: []any{nil}},
			},
		},
		{
			TestName: "Successfully assign chat to multiple groups",
			RequestBody: `{
				"groupIds": ["550e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-a716-446655440002"]
			}`,
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "AddChatToGroup", ExpectedResponse: []any{nil}},
			},
		},
		{
			TestName:       "Invalid JSON",
			RequestBody:    `{"invalid": json}`,
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName: "Invalid UUID in URI",
			RequestBody: `{
				"groupIds": ["550e8400-e29b-41d4-a716-446655440000"]
			}`,
			ChatId:         "invalid-chat-id",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName: "Invalid UUID in JSON",
			RequestBody: `{
				"groupIds": ["invalid-group-id"]
			}`,
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName: "Group manager error",
			RequestBody: `{
				"groupIds": ["550e8400-e29b-41d4-a716-446655440000"]
			}`,
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: []MockSetup{
				{Function: "AddChatToGroup", ExpectedResponse: []any{sharedErrors.ErrChatAlreadyInGroup}},
			},
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()

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
			mockGroupManager := mocks.NewMockGroupManager(t)

			for _, mc := range test.MockSetup {
				if mc.Function == "AddChatToGroup" {
					mockGroupManager.
						On("AddChatToGroup", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(mc.ExpectedResponse...).
						Maybe()
				}
			}

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/chats/"+test.ChatId+"/groups", bytes.NewBufferString(test.RequestBody))

			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			// Set URL parameters for the context
			ctx.Params = gin.Params{
				gin.Param{Key: "chatId", Value: test.ChatId},
			}

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
				mockGroupManager,
				tracer,
				mockMetricsServ,
			)

			controller.HandleAssignChatToGroups(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			if test.MockSetup != nil {
				mockGroupManager.AssertExpectations(t)
			}
		})
	}
}

//nolint:funlen
func TestHandleRemoveChatFromGroup(t *testing.T) {
	cfg, _ := config.LoadConfig()
	logger := slog.New(slog.DiscardHandler)
	tracer := otel.Tracer("test")

	type MockSetup struct {
		Function         string
		ExpectedResponse []any
	}

	tests := []struct {
		TestName       string
		ChatId         string
		GroupId        string
		ExpectedStatus int
		MockSetup      []MockSetup
	}{
		{
			TestName:       "Successfully remove chat from group",
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			GroupId:        "550e8400-e29b-41d4-a716-446655440000",
			ExpectedStatus: http.StatusOK,
			MockSetup: []MockSetup{
				{Function: "RemoveChatFromGroup", ExpectedResponse: []any{nil}},
			},
		},
		{
			TestName:       "Invalid chatId UUID",
			ChatId:         "invalid-chat-id",
			GroupId:        "550e8400-e29b-41d4-a716-446655440000",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName:       "Invalid groupId UUID",
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			GroupId:        "invalid-group-id",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup:      nil,
		},
		{
			TestName:       "Chat not in group error",
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			GroupId:        "550e8400-e29b-41d4-a716-446655440000",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: []MockSetup{
				{Function: "RemoveChatFromGroup", ExpectedResponse: []any{sharedErrors.ErrChatNotInGroup}},
			},
		},
		{
			TestName:       "Group manager internal error",
			ChatId:         "550e8400-e29b-41d4-a716-446655440001",
			GroupId:        "550e8400-e29b-41d4-a716-446655440000",
			ExpectedStatus: http.StatusBadRequest,
			MockSetup: []MockSetup{
				{Function: "RemoveChatFromGroup", ExpectedResponse: []any{errors.New("internal storage error")}},
			},
		},
	}

	mockMetricsServ := sharedMocks.NewMockMetricsService(t)
	mockMetricsServ.On("IncRequestSuccess").Return().Maybe()
	mockMetricsServ.On("IncRequestError", mock.Anything).Return().Maybe()
	mockMetricsServ.On("RecordRequestDuration", mock.Anything).Return().Maybe()

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
			mockGroupManager := mocks.NewMockGroupManager(t)

			for _, mc := range test.MockSetup {
				if mc.Function == "RemoveChatFromGroup" {
					mockGroupManager.
						On("RemoveChatFromGroup", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return(mc.ExpectedResponse...)
				}
			}

			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/chats/"+test.ChatId+"/groups/"+test.GroupId, nil)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req

			// Set URL parameters for the context
			ctx.Params = gin.Params{
				gin.Param{Key: "chatId", Value: test.ChatId},
				gin.Param{Key: "groupId", Value: test.GroupId},
			}

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
				mockGroupManager,
				tracer,
				mockMetricsServ,
			)

			controller.HandleRemoveChatFromGroup(ctx)

			if rec.Code != test.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", test.ExpectedStatus, rec.Code)
			}

			if test.MockSetup != nil {
				mockGroupManager.AssertExpectations(t)
			}
		})
	}
}
