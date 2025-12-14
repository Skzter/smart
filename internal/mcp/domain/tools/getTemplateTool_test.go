package tools

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/service"
)

//nolint:dupl,funlen
func TestNewGetTemplateTool(t *testing.T) {
	tests := []struct {
		name              string
		logger            *slog.Logger
		autotesterService *mocks.MockAutotesterAPIService
		expectedError     bool
		expectedErrorMsg  string
	}{
		{
			name:              "successful creation",
			logger:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
			autotesterService: mocks.NewMockAutotesterAPIService(t),
			expectedError:     false,
		},
		{
			name:              "nil logger",
			logger:            nil,
			autotesterService: mocks.NewMockAutotesterAPIService(t),
			expectedError:     true,
			expectedErrorMsg:  "nil",
		},
		{
			name:              "nil autotester service",
			logger:            slog.New(slog.NewTextHandler(os.Stdout, nil)),
			autotesterService: nil,
			expectedError:     true,
			expectedErrorMsg:  "nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool, err := NewGetTemplateTool(test.logger, test.autotesterService)

			if test.expectedError {
				assert.Error(t, err)
				assert.Nil(t, tool)
				if test.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), test.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tool)
				assert.Equal(t, test.logger, tool.logger)
				assert.Equal(t, test.autotesterService, tool.autotesterAPIService)
			}
		})
	}
}

// nolint:dupl nolint:funlen
func TestGetTemplateTool_GetTemplate(t *testing.T) {
	tests := []struct {
		name           string
		input          entity.TemplateIn
		mockSetup      func(*mocks.MockAutotesterAPIService)
		expectedError  bool
		expectedOutput entity.TemplateResponse
	}{
		{
			name:  "successful template retrieval",
			input: entity.TemplateIn{},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.TemplateResponse{
					Content: "describe('Test Suite', () => {\n  it('should pass', () => {\n    expect(true).toBe(true);\n  });\n});",
				}
				m.EXPECT().GetTemplate(mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.TemplateResponse{
				Content: "describe('Test Suite', () => {\n  it('should pass', () => {\n    expect(true).toBe(true);\n  });\n});",
			},
		},
		{
			name:  "service returns error",
			input: entity.TemplateIn{},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				m.EXPECT().GetTemplate(mock.Anything).
					Return(nil, errors.New("backend service unavailable")).Once()
			},
			expectedError:  true,
			expectedOutput: entity.TemplateResponse{},
		},
		{
			name:  "empty template content",
			input: entity.TemplateIn{},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				expectedResponse := &entity.TemplateResponse{
					Content: "",
				}
				m.EXPECT().GetTemplate(mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.TemplateResponse{
				Content: "",
			},
		},
		{
			name:  "large template content",
			input: entity.TemplateIn{},
			mockSetup: func(m *mocks.MockAutotesterAPIService) {
				largeContent := "import { test, expect } from '@playwright/test';\n" +
					"test.describe('Large Test Suite', () => {\n" +
					"  test('test 1', async ({ page }) => { /* ... */ });\n" +
					"  test('test 2', async ({ page }) => { /* ... */ });\n" +
					"});"
				expectedResponse := &entity.TemplateResponse{
					Content: largeContent,
				}
				m.EXPECT().GetTemplate(mock.Anything).
					Return(expectedResponse, nil).Once()
			},
			expectedError: false,
			expectedOutput: entity.TemplateResponse{
				Content: "import { test, expect } from '@playwright/test';\n" +
					"test.describe('Large Test Suite', () => {\n" +
					"  test('test 1', async ({ page }) => { /* ... */ });\n" +
					"  test('test 2', async ({ page }) => { /* ... */ });\n" +
					"});",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := mocks.NewMockAutotesterAPIService(t)
			test.mockSetup(mockService)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			tool, err := NewGetTemplateTool(logger, mockService)
			assert.NoError(t, err)

			ctx := context.Background()
			_, output, err := tool.GetTemplate(ctx, nil, test.input)

			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedOutput.Content, output.Content)
			}
		})
	}
}
