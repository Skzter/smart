package resource

import (
	"context"
	"log/slog"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/mcp/domain/mocks/store"
)

func TestNewTestLogStreamResource(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name        string
		logger      *slog.Logger
		store       *mocks.MockTestLogStreamStore
		expectError bool
		description string
	}{
		{
			name:        "valid initialization with all dependencies",
			logger:      logger,
			store:       mocks.NewMockTestLogStreamStore(t),
			expectError: false,
			description: "Should successfully create TestLogStreamResource when both logger and store are provided",
		},
		{
			name:        "nil logger returns error",
			logger:      nil,
			store:       mocks.NewMockTestLogStreamStore(t),
			expectError: true,
			description: "Should return an error when logger is nil",
		},
		{
			name:        "nil store returns error",
			logger:      logger,
			store:       nil,
			expectError: true,
			description: "Should return an error when store is nil",
		},
		{
			name:        "both dependencies nil returns error",
			logger:      nil,
			store:       nil,
			expectError: true,
			description: "Should return an error when both logger and store are nil",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resource, err := NewTestLogStreamResource(test.logger, test.store)

			if test.expectError {
				require.Error(t, err, test.description)
				require.Nil(t, resource, test.description)
			} else {
				require.NoError(t, err, test.description)
				require.NotNil(t, resource, test.description)
				require.Equal(t, test.logger, resource.logger, test.description)
				require.Equal(t, test.store, resource.store, test.description)
			}
		})
	}
}

func TestReadTestLogStream(t *testing.T) {
	logger := slog.Default()
	tests := []struct {
		name       string
		request    *mcp.ReadResourceRequest
		setupMocks func(*mocks.MockTestLogStreamStore)
		validate   func(*testing.T, *mcp.ReadResourceResult)
		expectErr  bool
	}{
		{
			name: "valid request returns formatted logs with events and metadata",
			request: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/test-123/logs",
				},
			},
			setupMocks: func(mockStore *mocks.MockTestLogStreamStore) {
				logStream := entity.NewLogStream()
				logStream.AddEvent(entity.LogEvent{Event: "log", Data: "Test started"})
				logStream.AddEvent(entity.LogEvent{Event: "log", Data: "Test completed"})
				logStream.SetComplete()

				mockStore.EXPECT().GetStream("test-123").Return(logStream, true)
			},
			validate: func(t *testing.T, result *mcp.ReadResourceResult) {
				require.NotNil(t, result)
				require.Len(t, result.Contents, 1)

				content := result.Contents[0]
				require.Equal(t, "mcp://tests/test-123/logs", content.URI)
				require.Equal(t, "application/json", content.MIMEType)

				expectedJSON := `{
								  "events": [
								    {
								      "Event": "log",
								      "Data": "Test started"
								    },
								    {
								      "Event": "log",
								      "Data": "Test completed"
								    }
								  ],
								  "meta": {
								    "final": true
								  }
								}`
				require.JSONEq(t, expectedJSON, content.Text)
			},
			expectErr: false,
		},
		{
			name: "invalid URI format returns nil and error",
			request: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/test-123/invalid",
				},
			},
			setupMocks: func(mockStore *mocks.MockTestLogStreamStore) {
				// Store should not be called for invalid URI
			},
			validate: func(t *testing.T, result *mcp.ReadResourceResult) {
				require.Nil(t, result)
			},
			expectErr: true,
		},
		{
			name: "stream does not exist returns nil and error",
			request: &mcp.ReadResourceRequest{
				Params: &mcp.ReadResourceParams{
					URI: "mcp://tests/non-existent-test/logs",
				},
			},
			setupMocks: func(mockStore *mocks.MockTestLogStreamStore) {
				mockStore.EXPECT().GetStream("non-existent-test").Return(nil, false)
			},
			validate: func(t *testing.T, result *mcp.ReadResourceResult) {
				require.Nil(t, result)
			},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStore := mocks.NewMockTestLogStreamStore(t)
			test.setupMocks(mockStore)

			resource, err := NewTestLogStreamResource(logger, mockStore)
			require.NoError(t, err)
			require.NotNil(t, resource)

			result, err := resource.ReadTestLogStream(context.Background(), test.request)

			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			test.validate(t, result)
		})
	}
}
