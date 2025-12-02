package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	autoRepoMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/repository"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
)

func TestNewDocker(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tests := []struct {
		name    string
		logger  *slog.Logger
		config  *config.Config
		fs      repository.LogFileSystem
		client  DockerClient
		wantErr bool
	}{
		{
			name:    "success - valid docker service",
			logger:  logger,
			config:  cfg,
			fs:      autoRepoMocks.NewMockFileSystem(t),
			client:  mocks.NewMockDockerClient(t),
			wantErr: false,
		},
		{
			name:    "error - nil values",
			logger:  nil,
			config:  nil,
			fs:      nil,
			client:  nil,
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serv, err := NewDocker(tc.logger, tc.config, tc.fs, tc.client)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, serv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, serv)
			}
		})
	}
}

func TestReadLog(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()

	tests := []struct {
		name             string
		filename         string
		mockResponseRead []any
		expectError      bool
	}{
		{
			name:             "success - reads log without error",
			filename:         "test1",
			mockResponseRead: []any{[]byte("content"), nil},
			expectError:      false,
		},
		{
			name:             "failure - log reading fails",
			filename:         "test1",
			mockResponseRead: []any{nil, errors.New("failure to read")},
			expectError:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockLogfilesystem := autoRepoMocks.NewMockFileSystem(t)
			mockDockerClient := mocks.NewMockDockerClient(t)
			if tc.mockResponseRead != nil {
				mockLogfilesystem.On("ReadFile", mock.Anything).Return(tc.mockResponseRead...)
			}
			dockerServ, _ := NewDocker(logger, cfg, mockLogfilesystem, mockDockerClient)
			content, err := dockerServ.ReadLog(tc.filename)
			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

// nolint:funlen
func TestRunTest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	fs := autoRepoMocks.NewMockFileSystem(t)

	tests := []struct {
		name               string
		ctx                context.Context
		filename           string
		mockResponseCreate []any
		mockResponseStart  []any
		mockResponseWait   []any
		wantErr            bool
	}{
		{
			name:     "error - nil ctx",
			ctx:      nil,
			filename: "test.log",
			wantErr:  true,
		},
		{
			name:     "error - empty filename",
			ctx:      t.Context(),
			filename: "",
			wantErr:  true,
		},
		{
			name:     "error - container creation fails",
			ctx:      t.Context(),
			filename: "test.log",
			mockResponseCreate: []any{container.CreateResponse{
				ID:       "1",
				Warnings: []string{"not safe"},
			}, errors.New("failure in container creation")},
			wantErr: true,
		},
		{
			name:     "error - container starting fails",
			ctx:      t.Context(),
			filename: "test.log",
			mockResponseCreate: []any{container.CreateResponse{
				ID:       "1",
				Warnings: nil,
			}, nil},
			mockResponseStart: []any{errors.New("failure when starting container")},
			wantErr:           true,
		},
		{
			name:     "error - waiting for container fails",
			ctx:      t.Context(),
			filename: "test.log",
			mockResponseCreate: []any{container.CreateResponse{
				ID:       "1",
				Warnings: nil,
			}, nil},
			mockResponseStart: []any{nil},
			mockResponseWait:  []any{nil, createErrorChannel(errors.New("error in channel"))},
			wantErr:           true,
		},
		{
			name:     "error - container exits with non zero stauscode",
			ctx:      t.Context(),
			filename: "test.log",
			mockResponseCreate: []any{container.CreateResponse{
				ID:       "1",
				Warnings: nil,
			}, nil},
			mockResponseStart: []any{nil},
			mockResponseWait: []any{createStatusChannel(container.WaitResponse{
				Error: &container.WaitExitError{
					Message: "something went wrong",
				},
				StatusCode: 1,
			}), nil},
			wantErr: true,
		},
		{
			name:     "successful - container executes without errors",
			ctx:      t.Context(),
			filename: "test.log",
			mockResponseCreate: []any{container.CreateResponse{
				ID:       "1",
				Warnings: nil,
			}, nil},
			mockResponseStart: []any{nil},
			mockResponseWait: []any{createStatusChannel(container.WaitResponse{
				StatusCode: 0,
			}), createErrorChannel(nil)},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockDockerClient := mocks.NewMockDockerClient(t)
			t.Logf("TESTCASE:\nMockCreate: %v\nMockStart: %v\nMockWait: %v\n", tc.mockResponseCreate, tc.mockResponseStart, tc.mockResponseWait)
			if tc.mockResponseCreate != nil {
				mockDockerClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mockResponseCreate...)
			}
			if tc.mockResponseStart != nil {
				mockDockerClient.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(tc.mockResponseStart...)
			}
			if tc.mockResponseWait != nil {
				mockDockerClient.On("ContainerWait", mock.Anything, mock.Anything, mock.Anything).Return(tc.mockResponseWait...)
			}
			dockerServ := &docker{
				logger:     logger,
				config:     cfg,
				filesystem: fs,
				client:     mockDockerClient,
			}
			err := dockerServ.RunTest(tc.ctx, tc.filename)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createErrorChannel(err error) <-chan error {
	errorC := make(chan error, 1) // Create bidirectional channel
	errorC <- err
	return errorC // Automatically converts to receive-only on return
}

func createStatusChannel(resp container.WaitResponse) <-chan container.WaitResponse {
	respC := make(chan container.WaitResponse, 1)
	respC <- resp
	return respC
}
