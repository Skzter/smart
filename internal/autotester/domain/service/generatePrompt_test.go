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
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	autotesterMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/mocks/service"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
)

// nolint: dupl
func TestNewGeneratePromptService(t *testing.T) {
	openai := mocks.NewMockOpenAI(t)
	taglist := mocks.NewMockTaglistStorage(t)
	logger := slog.Default()
	cfg := config.Config{}
	validator := autotesterMocks.NewMockValidator(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		openai    srv.OpenAI
		taglist   srv.TaglistStorage
		config    *config.Config
		logger    *slog.Logger
		validator Validator
		wantErr   bool
	}{
		{
			name:      "all not nil",
			openai:    openai,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: validator,
			wantErr:   false,
		},
		{
			name:      "nil openai",
			openai:    nil,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil config",
			openai:    openai,
			taglist:   taglist,
			config:    nil,
			logger:    logger,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil logger",
			openai:    openai,
			taglist:   taglist,
			config:    &cfg,
			logger:    nil,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil validator",
			openai:    openai,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: nil,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewGeneratePromptService(test.openai, test.taglist, test.config, test.logger, test.validator, tracer)
			if test.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, repo)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestGeneratePrompt(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		Prompts: &config.Prompts{
			AutoPlaywrightPromptT: "system prompt %s",
		},
	}
	tags := &sharedEntity.TagList{Tags: []sharedEntity.Tag{{Name: "Tag1", Description: ""}, {Name: "Tag2", Description: ""}}}
	code := "some code"
	tracer := otel.Tracer("test")

	tests := []struct {
		name              string
		expectedResult    string
		validateReturns   []any
		requestReturns    []any
		getTaglistReturns []any
		expectErr         bool
		ctx               context.Context
	}{
		{
			name:              "success",
			expectedResult:    code,
			validateReturns:   []any{nil},
			requestReturns:    []any{&sharedEntity.Message{Role: "assistant", Body: code}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         false,
			ctx:               context.Background(),
		},
		{
			name:              "getTaglist error",
			expectedResult:    code,
			validateReturns:   []any{nil},
			requestReturns:    []any{&sharedEntity.Message{Role: "assistant", Body: code}, nil},
			getTaglistReturns: []any{nil, errors.New("err")},
			expectErr:         false,
			ctx:               context.Background(),
		},
		{
			name:      "nil ctx",
			ctx:       nil,
			expectErr: true,
		},
		{
			name:              "validate error",
			expectedResult:    code,
			validateReturns:   []any{errors.New("err")},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
		{
			name:              "empty code segment in openau response",
			validateReturns:   []any{nil},
			requestReturns:    []any{&sharedEntity.Message{Role: "assistant", Body: ""}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
		{
			name:              "openai error",
			validateReturns:   []any{nil},
			requestReturns:    []any{nil, sharedErrors.ErrInternalServer},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openai := mocks.NewMockOpenAI(t)
			taglist := mocks.NewMockTaglistStorage(t)
			validator := autotesterMocks.NewMockValidator(t)

			if tt.validateReturns != nil {
				validator.On("ValidateRequest", mock.Anything, mock.Anything).Return(tt.validateReturns...)
			}

			if tt.requestReturns != nil {
				openai.On("Request", mock.Anything, mock.Anything).Return(tt.requestReturns...)
			}

			if tt.getTaglistReturns != nil {
				taglist.On("GetTaglist", mock.Anything).Return(tt.getTaglistReturns...)
			}

			svc, _ := NewGeneratePromptService(openai, taglist, cfg, logger, validator, tracer)
			got, err := svc.GeneratePrompt(tt.ctx, &entity.Chat{}, &entity.UserRequest{})

			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Empty(t, got)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedResult, got)
			}
		})
	}
}
