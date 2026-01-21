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
	"go.opentelemetry.io/otel"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	mocks "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/mocks/service"
)

func TestNewDocker(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tracer := otel.Tracer("test")

	tests := []struct {
		name    string
		logger  *slog.Logger
		config  *config.Autotester
		client  DockerClient
		wantErr bool
	}{
		{
			name:    "success",
			logger:  logger,
			config:  cfg,
			client:  mocks.NewMockDockerClient(t),
			wantErr: false,
		},
		{
			name:    "error - nil arguments",
			logger:  nil,
			config:  nil,
			client:  nil,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewDocker(tc.logger, tc.config, tc.client, tracer)

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
	tracer := otel.Tracer("test")

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

			// mock ContainerCreate
			if tc.createResp != nil {
				mockClient.On("ContainerCreate",
					mock.Anything,
					mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(tc.createResp...)
			}

			// mock ContainerStart
			if tc.startResp != nil {
				mockClient.On("ContainerStart",
					mock.Anything, "123", mock.Anything,
				).Return(tc.startResp...)
			}

			d := &docker{
				logger:           logger,
				config:           cfg,
				client:           mockClient,
				testContainerMap: make(map[string]*entity.ContainerInfo),
				tracer:           tracer,
			}

			id, err := d.RunTest(tc.ctx, tc.filename, tc.testID, "userX", "chatY")

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "123", id)

				info, ok := d.testContainerMap[tc.testID]
				assert.True(t, ok)
				assert.Equal(t, "123", info.ContainerID)
				assert.Equal(t, "userX", info.UserID)
				assert.Equal(t, "chatY", info.ChatID)
			}
		})
	}
}

func TestAttachToContainer(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()

	mockClient := mocks.NewMockDockerClient(t)
	expectedValue := types.HijackedResponse{}

	mockClient.On("ContainerAttach",
		mock.Anything, "123", mock.Anything,
	).Return(expectedValue, nil)

	d := &docker{
		logger:           logger,
		config:           cfg,
		client:           mockClient,
		testContainerMap: make(map[string]*entity.ContainerInfo),
	}

	resp, err := d.AttachToContainer(context.Background(), "123")

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, *resp)
}

func TestWaitContainer(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	mockClient := mocks.NewMockDockerClient(t)

	status := make(chan container.WaitResponse)
	errs := make(chan error)

	statusCh := (<-chan container.WaitResponse)(status)
	errCh := (<-chan error)(errs)

	mockClient.On("ContainerWait",
		mock.Anything, "123", mock.Anything,
	).Return(statusCh, errCh)

	d := &docker{
		logger:           logger,
		config:           cfg,
		client:           mockClient,
		testContainerMap: make(map[string]*entity.ContainerInfo),
	}

	s, e := d.WaitContainer(context.Background(), "123")

	assert.Equal(t, statusCh, s)
	assert.Equal(t, errCh, e)
}

func TestGetContainerInfo(t *testing.T) {
	d := &docker{
		testContainerMap: map[string]*entity.ContainerInfo{
			"t1": {ContainerID: "cid1", UserID: "u1", ChatID: "s1"},
		},
	}

	info, found := d.GetContainerInfo("t1")
	assert.True(t, found)
	assert.Equal(t, "cid1", info.ContainerID)

	_, found = d.GetContainerInfo("missing")
	assert.False(t, found)
}
