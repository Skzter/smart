package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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

// MockChatModel is a mock implementation of model.ChatModel
type MockChatModel struct {
	mock.Mock
}

func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// The mock.Called method doesn't handle variadic arguments well when matching.
	// We can pass the variadic args slice as a single argument if needed, or just ignore them for simple tests.
	// Here passing generic "mock.Anything" for opts usually works if we don't care.
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.Message), args.Error(1)
}

func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*schema.StreamReader[*schema.Message]), args.Error(1)
}

func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	args := m.Called(tools)
	return args.Error(0)
}

// nolint: dupl
func TestNewGeneratePromptService(t *testing.T) {
	chatModel := &MockChatModel{}
	taglist := mocks.NewMockTaglistStorage(t)
	logger := slog.Default()
	cfg := config.Config{}
	validator := autotesterMocks.NewMockValidator(t)
	tracer := otel.Tracer("test")

	tests := []struct {
		name      string
		chatModel model.ChatModel
		taglist   srv.TaglistStorage
		config    *config.Config
		logger    *slog.Logger
		validator Validator
		wantErr   bool
	}{
		{
			name:      "all not nil",
			chatModel: chatModel,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: validator,
			wantErr:   false,
		},
		{
			name:      "nil chatModel",
			chatModel: nil,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil config",
			chatModel: chatModel,
			taglist:   taglist,
			config:    nil,
			logger:    logger,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil logger",
			chatModel: chatModel,
			taglist:   taglist,
			config:    &cfg,
			logger:    nil,
			validator: validator,
			wantErr:   true,
		},
		{
			name:      "nil validator",
			chatModel: chatModel,
			taglist:   taglist,
			config:    &cfg,
			logger:    logger,
			validator: nil,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewGeneratePromptService(test.chatModel, test.taglist, test.config, test.logger, test.validator, tracer)
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
		generateReturns   []any
		getTaglistReturns []any
		expectErr         bool
		ctx               context.Context
	}{
		{
			name:              "success",
			expectedResult:    code,
			validateReturns:   []any{nil},
			generateReturns:   []any{&schema.Message{Content: code, Role: schema.Assistant}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         false,
			ctx:               context.Background(),
		},
		{
			name:              "getTaglist error",
			expectedResult:    code,
			validateReturns:   []any{nil},
			generateReturns:   []any{&schema.Message{Content: code, Role: schema.Assistant}, nil},
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
			generateReturns:   []any{&schema.Message{Content: "", Role: schema.Assistant}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
		{
			name:              "openai error",
			validateReturns:   []any{nil},
			generateReturns:   []any{nil, sharedErrors.ErrInternalServer},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chatModel := &MockChatModel{}
			taglist := mocks.NewMockTaglistStorage(t)
			validator := autotesterMocks.NewMockValidator(t)

			if tt.validateReturns != nil {
				validator.On("ValidateRequest", mock.Anything, mock.Anything).Return(tt.validateReturns...)
			}

			if tt.generateReturns != nil {
				// Eino calls Generate with (ctx, input)
				chatModel.On("Generate", mock.Anything, mock.Anything).Return(tt.generateReturns...)
			}

			if tt.getTaglistReturns != nil {
				taglist.On("GetTaglist", mock.Anything).Return(tt.getTaglistReturns...)
			}

			svc, _ := NewGeneratePromptService(chatModel, taglist, cfg, logger, validator, tracer)
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
