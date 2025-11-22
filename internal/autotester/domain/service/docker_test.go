package service

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	autoRepoMocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository/mocks"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service/mocks"
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
