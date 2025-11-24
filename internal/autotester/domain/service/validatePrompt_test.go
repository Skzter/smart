package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

//nolint:dupl
func TestNewValidatePromptService(t *testing.T) {
	service := mocks.NewMockOpenAI(t)
	logger := slog.Default()
	cfg := config.Config{}
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		service srv.OpenAI
		config  *config.Config
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			service: service,
			config:  &cfg,
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil service",
			service: nil,
			config:  &cfg,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil config",
			service: service,
			config:  nil,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil logger",
			service: service,
			config:  &cfg,
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewValidatePromptService(test.service, test.config, test.logger, tracer)
			if test.wantErr {
				assert.Nil(t, repo)
				assert.NotNil(t, err)
			}
			if !test.wantErr && repo == nil {
				assert.NotNil(t, repo)
				assert.Nil(t, err)
			}
		})
	}
}

//nolint:funlen
func TestValidatePrompt(t *testing.T) {
	serviceMock := mocks.NewMockOpenAI(t)
	cfg := &config.Config{
		Model: "",
		Prompts: &config.Prompts{
			ValidationPrompt: "",
		},
	}
	logger := slog.New(slog.DiscardHandler)
	svc, _ := NewValidatePromptService(serviceMock, cfg, logger)

	validPrompt := "valid Prompt"
	invalidPrompt := "invalid Prompt"
	invalidJson := "invalid json"

	validRequest := entity.Request{
		Messages: []entity.Message{{Role: "user", Body: validPrompt}},
	}
	validResponse := entity.Message{
		Body: `{"valid": true, "message": ""}`,
	}

	invalidRequest := entity.Request{
		Messages: []entity.Message{{Role: "user", Body: invalidPrompt}},
	}
	invalidResponse := entity.Message{
		Body: `{"valid": false, "message": "alle gründe warum es schiefgelaufen"}`,
	}

	invalidJsonRequest := entity.Request{
		Messages: []entity.Message{{Role: "user", Body: invalidJson}},
	}
	invalidJsonResponse := entity.Message{
		Body: "no json",
	}

	mockSetup := []struct {
		request     entity.Request
		response    *entity.Message
		returnError error
		wantErr     bool
	}{
		{
			request:     validRequest,
			response:    &validResponse,
			returnError: nil,
		},
		{
			request:     invalidRequest,
			response:    &invalidResponse,
			returnError: nil,
		},
		{
			request:     invalidJsonRequest,
			response:    &invalidJsonResponse,
			returnError: nil,
		},
		{
			request: entity.Request{
				Messages: []entity.Message{{Role: "user", Body: ""}},
			},
			response:    nil,
			returnError: sharedErrors.ErrInternalServer,
		},
	}
	tracer := otel.Tracer("test")

	tests := []struct {
		name        string
		userPrompt  string
		ctx         context.Context
		wantErr     bool
		isValid     bool
		expectedErr error
	}{
		{
			name:        "valid request without changes",
			userPrompt:  validPrompt,
			ctx:         context.Background(),
			wantErr:     false,
			isValid:     true,
			expectedErr: nil,
		},
		{
			name:        "valid request but with changes",
			userPrompt:  invalidPrompt,
			ctx:         context.Background(),
			wantErr:     false,
			isValid:     false,
			expectedErr: nil,
		},
		{
			name:        "invalid json",
			userPrompt:  invalidJson,
			ctx:         context.Background(),
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrInternalServer,
		},
		{
			name:        "nil ctx",
			userPrompt:  validPrompt,
			ctx:         nil,
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrInternalServer,
		},
		{
			name:        "service error",
			userPrompt:  "",
			ctx:         context.Background(),
			wantErr:     true,
			isValid:     false,
			expectedErr: sharedErrors.ErrValidation,
		},
	}

	for _, mc := range mockSetup {
		serviceMock.On("Request", mock.Anything, mc.request).Return(mc.response, mc.returnError)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceMock := mocks.NewMockOpenAI(t)
			rec := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(rec)

			tt.mockSetup((*mocks.MockOpenAI)(serviceMock), ctx)

			svc := &validatePrompt{
				service: serviceMock,
				config:  cfg,
				logger:  logger,
				tracer:  tracer,
			}
			valid, str, err := svc.ValidatePrompt(ctx, tt.userPrompt)

			if tt.wantErr {
				assert.NotNil(t, err)
				if !errors.Is(err, tt.expectedErr) {
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.Nil(t, err)
				if tt.isValid {
					assert.Equal(t, tt.isValid, valid)
					assert.Empty(t, str)
				} else {
					assert.NotEmpty(t, str)
				}
			}
			mock.AssertExpectationsForObjects(t)
		})
	}
}
