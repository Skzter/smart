package service

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/docker/docker/api/types"
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
			name:    "success",
			logger:  logger,
			config:  cfg,
			fs:      autoRepoMocks.NewMockFileSystem(t),
			client:  mocks.NewMockDockerClient(t),
			wantErr: false,
		},
		{
			name:    "error - nil arguments",
			logger:  nil,
			config:  nil,
			fs:      nil,
			client:  nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewDocker(tc.logger, tc.config, tc.fs, tc.client)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, srv)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, srv)
			}
		})
	}
}
func TestRunTest(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	fs := autoRepoMocks.NewMockFileSystem(t)

	tests := []struct {
		name       string
		filename   string
		testID     string
		ctx        context.Context
		createResp []any
		startResp  []any
		wantErr    bool
	}{
		{
			name:     "error - container create fails",
			filename: "/tmp/test.spec.js",
			testID:   "abc",
			ctx:      context.Background(),
			createResp: []any{
				container.CreateResponse{}, errors.New("create error"),
			},
			wantErr: true,
		},
		{
			name:     "error - container start fails",
			filename: "/tmp/test.spec.js",
			testID:   "abc",
			ctx:      context.Background(),
			createResp: []any{
				container.CreateResponse{ID: "123"}, nil,
			},
			startResp: []any{
				errors.New("start error"),
			},
			wantErr: true,
		},
		{
			name:     "success - container started",
			filename: "/tmp/test.spec.js",
			testID:   "abc",
			ctx:      context.Background(),
			createResp: []any{
				container.CreateResponse{ID: "123"}, nil,
			},
			startResp: []any{nil},
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mocks.NewMockDockerClient(t)

			if tc.createResp != nil {
				mockClient.On("ContainerCreate",
					mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(tc.createResp...)
			}

			if tc.startResp != nil {
				mockClient.On("ContainerStart",
					mock.Anything, "123", mock.Anything,
				).Return(tc.startResp...)
			}

			d := &docker{
				logger:           logger,
				config:           cfg,
				filesystem:       fs,
				client:           mockClient,
				testContainerMap: make(map[string]string),
			}

			id, err := d.RunTest(tc.ctx, tc.filename, tc.testID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "123", id)
				assert.Equal(t, "123", d.testContainerMap[tc.testID])
			}
		})
	}
}
func TestAttachToContainer(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()

	mockFilesystem := autoRepoMocks.NewMockFileSystem(t)
	mockClient := mocks.NewMockDockerClient(t)

	expectedValue := types.HijackedResponse{
		Conn:   nil,
		Reader: nil,
	}

	mockClient.On("ContainerAttach",
		mock.Anything, "123", mock.Anything,
	).Return(expectedValue, nil)

	d := &docker{
		logger:           logger,
		config:           cfg,
		filesystem:       mockFilesystem,
		client:           mockClient,
		testContainerMap: make(map[string]string),
	}

	resp, err := d.AttachToContainer(context.Background(), "123")

	// === Assertions ===
	assert.NoError(t, err)

	assert.Equal(t, expectedValue, *resp)
}

func TestWaitContainer(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()

	mockFilesystem := autoRepoMocks.NewMockFileSystem(t)
	mockClient := mocks.NewMockDockerClient(t)

	// echte Channels
	status := make(chan container.WaitResponse)
	errs := make(chan error)

	// cast zu receive-only (GENAU das erwartet mockery)
	statusCh := (<-chan container.WaitResponse)(status)
	errCh := (<-chan error)(errs)

	mockClient.On("ContainerWait",
		mock.Anything, "123", mock.Anything,
	).Return(statusCh, errCh)

	d := &docker{
		logger:           logger,
		config:           cfg,
		filesystem:       mockFilesystem,
		client:           mockClient,
		testContainerMap: make(map[string]string),
	}

	s, e := d.WaitContainer(context.Background(), "123")

	assert.NotNil(t, s)
	assert.NotNil(t, e)

	// optional: pointer compare
	assert.Equal(t, statusCh, s)
	assert.Equal(t, errCh, e)
}

func TestGetContainerID(t *testing.T) {
	d := &docker{
		testContainerMap: map[string]string{"t1": "cid123"},
	}

	id, found := d.GetContainerID("t1")
	assert.True(t, found)
	assert.Equal(t, "cid123", id)

	_, found = d.GetContainerID("missing")
	assert.False(t, found)
}
