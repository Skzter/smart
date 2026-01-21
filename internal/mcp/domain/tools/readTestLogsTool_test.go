package tools

import (
	"context"
	"log/slog"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/store"
)

func TestNewReadTestLogStreamTool(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		logger      *slog.Logger
		store       *mocks.MockTestLogStreamStore
		expectError bool
	}{
		{
			name:        "valid initialization with all dependencies",
			logger:      logger,
			store:       mocks.NewMockTestLogStreamStore(t),
			expectError: false,
		},
		{
			name:        "nil logger returns error",
			logger:      nil,
			store:       mocks.NewMockTestLogStreamStore(t),
			expectError: true,
		},
		{
			name:        "nil store returns error",
			logger:      logger,
			store:       nil,
			expectError: true,
		},
		{
			name:        "both dependencies nil returns error",
			logger:      nil,
			store:       nil,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tool, err := NewReadTestLogsTool(test.logger, test.store)

			if test.expectError {
				require.Error(t, err)
				require.Nil(t, tool)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tool)
				require.Equal(t, test.logger, tool.logger)
				require.Equal(t, test.store, tool.store)
			}
		})
	}
}

func TestReadTestLogStream(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name       string
		input      entity.ReadTestLogStreamIn
		setupMocks func(*mocks.MockTestLogStreamStore)
		validate   func(*testing.T, *mcp.CallToolResult, *entity.ReadTestLogStreamResponse)
		expectErr  bool
	}{
		{
			name: "read existing test logs returns events and final metadata",
			input: entity.ReadTestLogStreamIn{
				TestId: "test-456",
			},
			setupMocks: func(mockStore *mocks.MockTestLogStreamStore) {
				logStream := entity.NewLogStream()
				logStream.AddEvent(entity.LogEvent{Event: "log", Data: "Test running"})
				logStream.AddEvent(entity.LogEvent{Event: "log", Data: "Test passed"})
				logStream.SetComplete()

				mockStore.EXPECT().GetStream("test-456").Return(logStream, true)
			},
			validate: func(t *testing.T, result *mcp.CallToolResult, output *entity.ReadTestLogStreamResponse) {
				require.Nil(t, result)
				require.NotNil(t, output)
				require.Len(t, output.Content, 2)
				require.Equal(t, "log", output.Content[0].Event)
				require.Equal(t, "Test running", output.Content[0].Data)
				require.Equal(t, "log", output.Content[1].Event)
				require.Equal(t, "Test passed", output.Content[1].Data)

				require.Equal(t, true, output.Meta["final"])
			},
			expectErr: false,
		},
		{
			name: "read non-existing test returns empty content and exists false",
			input: entity.ReadTestLogStreamIn{
				TestId: "non-existent-test",
			},
			setupMocks: func(mockStore *mocks.MockTestLogStreamStore) {
				mockStore.EXPECT().GetStream("non-existent-test").Return(nil, false)
			},
			validate: func(t *testing.T, result *mcp.CallToolResult, output *entity.ReadTestLogStreamResponse) {
				require.Nil(t, result)
				require.NotNil(t, output)
				require.Empty(t, output.Content)
				require.Equal(t, false, output.Meta["exists"])
			},
			expectErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := mocks.NewMockTestLogStreamStore(t)
			test.setupMocks(mockStore)

			tool, err := NewReadTestLogsTool(logger, mockStore)
			require.NoError(t, err)
			require.NotNil(t, tool)

			result, output, err := tool.ReadTestLogStream(context.Background(), nil, test.input)

			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			test.validate(t, result, output)
		})
	}
}
