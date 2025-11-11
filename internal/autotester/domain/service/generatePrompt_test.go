package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	srv "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/service/mocks"
)

// nolint: dupl
func TestNewGeneratePromptService(t *testing.T) {
	openai := mocks.NewMockOpenAI(t)
	taglist := mocks.NewMockTaglistStorage(t)
	logger := slog.Default()
	cfg := config.Config{}

	tests := []struct {
		name    string
		openai  srv.OpenAI
		taglist srv.TaglistStorage
		config  *config.Config
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "all not nil",
			openai:  openai,
			taglist: taglist,
			config:  &cfg,
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "nil openai",
			openai:  nil,
			taglist: taglist,
			config:  &cfg,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil taglist",
			openai:  openai,
			taglist: nil,
			config:  &cfg,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil config",
			openai:  openai,
			taglist: taglist,
			config:  nil,
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil logger",
			openai:  openai,
			taglist: taglist,
			config:  &cfg,
			logger:  nil,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := NewGeneratePromptService(test.openai, test.taglist, test.config, test.logger)
			if (err != nil) != test.wantErr {
				t.Errorf("NewGeneratePromptService() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && repo == nil {
				t.Errorf("NewGeneratePromptService() returned nil service")
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
	tags := []string{"Tag1, Tag2"}
	code := "some code"

	tests := []struct {
		name              string
		expectedResult    string
		requestReturns    []any
		getTaglistReturns []any
		expectErr         bool
		ctx               context.Context
	}{
		{
			name:              "success",
			expectedResult:    code,
			requestReturns:    []any{&sharedEntity.Response{Text: code}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         false,
			ctx:               context.Background(),
		},
		{
			name:      "nil ctx",
			ctx:       nil,
			expectErr: true,
		},
		{
			name:              "taglist error",
			getTaglistReturns: []any{[]string{}, errors.New("Err")},
			expectErr:         true,
			ctx:               context.Background(),
		},
		{
			name:              "empty code segment in openau response",
			requestReturns:    []any{&sharedEntity.Response{Text: ""}, nil},
			getTaglistReturns: []any{tags, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
		{
			name:              "openai error",
			requestReturns:    []any{nil, sharedErrors.ErrInternalServer},
			getTaglistReturns: []any{[]string{"Tag1, Tag2"}, nil},
			expectErr:         true,
			ctx:               context.Background(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openai := mocks.NewMockOpenAI(t)
			taglist := mocks.NewMockTaglistStorage(t)

			//TODO: refactor to reduce duplication with other tests
			if tt.requestReturns != nil {
				openai.On("Request", mock.Anything, mock.Anything).Return(tt.requestReturns...)
			}

			if tt.getTaglistReturns != nil {
				taglist.On("GetTaglist", mock.Anything).Return(tt.getTaglistReturns...)
			}

			svc, _ := NewGeneratePromptService(openai, taglist, cfg, logger)
			got, err := svc.GeneratePrompt(tt.ctx, "user says hi", "session-123")

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
