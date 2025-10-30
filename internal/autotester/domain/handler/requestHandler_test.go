package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errs"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
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
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			controller, err := NewAutotesterController(test.logger, test.config, mockValServ, mockGenServ)

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

	jsonErr := errors.New("json error")
	customJsonErr := &errs.Error{
		Message:    "invalid json",
		Underlying: jsonErr,
		Type:       errs.Private,
	}
	mockSetup := []struct {
		function         string
		userPrompt       string
		sessionID        string
		expectedResponse string
		ResponseError    error
	}{
		{
			function:         "ValidatePrompt",
			userPrompt:       validPrompt,
			sessionID:        sessionid,
			expectedResponse: "",
			ResponseError:    nil,
		},
		{
			function:         "GeneratePrompt",
			userPrompt:       validPrompt,
			sessionID:        sessionid,
			expectedResponse: "This is a generated Prompt",
			ResponseError:    nil,
		},
		{
			// no need for generate mock
			function:         "ValidatePrompt",
			userPrompt:       invalidPrompt,
			sessionID:        sessionid,
			expectedResponse: "versuch doch mal das",
			ResponseError:    nil,
		},
		{
			// errors in validation
			function:         "ValidatePrompt",
			userPrompt:       "json gibts nicht",
			sessionID:        sessionid,
			expectedResponse: "",
			ResponseError:    customJsonErr,
		},
		{
			// test has to pass in validation in order to fail in generation below
			function:         "ValidatePrompt",
			userPrompt:       "generating err",
			sessionID:        sessionid,
			expectedResponse: "",
			ResponseError:    nil,
		},
		{
			function:         "GeneratePrompt",
			userPrompt:       "generating err",
			sessionID:        sessionid,
			expectedResponse: "",
			ResponseError:    errs.ErrEmptyResponse,
		},
		{
			// test for frontend facing error but currently none available
			function:         "ValidatePrompt",
			userPrompt:       "validation failure results in fe error",
			sessionID:        sessionid,
			expectedResponse: "",
			ResponseError: &errs.Error{
				Underlying: errors.New("frontend error"),
				Type:       errs.Public,
			},
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
			ExpectedStatus: http.StatusOK,
		},
		{
			TestName: "valid request, validate will return false json",
			RequestBody: `{
				"message": {
					"data":  "json gibts nicht",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			TestName: "valid request, errors when generating",
			RequestBody: `{
				"message": {
					"data":  "generating err",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusInternalServerError,
		},
		{
			TestName: "invalid request, user error",
			RequestBody: `{
				"message": {
					"data":  "validation failure results in fe error",
					"agent": "user"
				},
				"userId":         "2",
				"conversationId": "2"
			}`,
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	mockGenServ := mocks.NewMockGeneratePrompt(t)
	mockValServ := mocks.NewMockValidatePrompt(t)

	// setup mocks
	for _, mc := range mockSetup {
		if mc.function == "ValidatePrompt" {
			mockValServ.On(mc.function, mock.Anything, mc.userPrompt, mc.sessionID).Return(mc.expectedResponse, mc.ResponseError)
		}
		if mc.function == "GeneratePrompt" {
			mockGenServ.On(mc.function, mock.Anything, mc.userPrompt, mc.sessionID).Return(mc.expectedResponse, mc.ResponseError)
		}
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/chat", bytes.NewBufferString(test.RequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Errors.Errors()

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ)

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

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/v1/userInfo", bytes.NewBufferString(test.UserRequestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)
			ctx.Request = req
			ctx.Errors.Errors()

			controller, err := NewAutotesterController(logger, cfg, mockValServ, mockGenServ)

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

func TestHandleError(t *testing.T) {
	errPublic := errors.New("public error")
	publicCustomErr := &errs.Error{
		Message:    fmt.Sprintf("wild error: %v", errPublic),
		Underlying: errPublic,
		Type:       errs.Public,
	}

	tests := []struct {
		name         string
		givenError   error
		wantedError  error
		wantedStatus int
		wantErr      bool
	}{
		{
			name:         "nil error",
			givenError:   nil,
			wantedError:  nil,
			wantedStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "custom: repo error: empty response",
			givenError:   errs.ErrEmptyResponse,
			wantedError:  errs.ErrInternalServer,
			wantedStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "custom: service error: nil ctx",
			givenError:   &assert.NotNilError{Message: "assert failed"},
			wantedError:  errs.ErrInternalServer,
			wantedStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "connection error",
			givenError:   errors.New("Post failure to connect to api.openai.com"),
			wantedError:  errs.ErrInternalServer,
			wantedStatus: http.StatusInternalServerError,
			wantErr:      true,
		},
		{
			name:         "custom: public error",
			givenError:   publicCustomErr,
			wantedError:  publicCustomErr,
			wantedStatus: http.StatusBadRequest,
			wantErr:      true,
		},
		{
			name:         "public error",
			givenError:   errPublic,
			wantedError:  errPublic,
			wantedStatus: http.StatusBadRequest,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := handleError(tt.givenError)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleError() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !errors.Is(err, tt.wantedError) {
				t.Errorf("gave back wrong error: got: %v, wanted error: %v", err, tt.wantedError)
			}
			if status != tt.wantedStatus {
				t.Errorf("gave back wrong code: got %d, wanted %d", status, tt.wantedStatus)
			}
		})
	}
}
