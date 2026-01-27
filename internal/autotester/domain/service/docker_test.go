package service

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
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

//nolint:funlen
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
		waitResp   []any
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
			waitResp:  []any{},
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mocks.NewMockDockerClient(t)

			if tc.createResp != nil {
				mockClient.On("ContainerCreate",
					mock.Anything,
					mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(tc.createResp...)
			}

			if tc.startResp != nil {
				mockClient.On("ContainerStart",
					mock.Anything, "123", mock.Anything,
				).Return(tc.startResp...)
			}

			if tc.waitResp != nil {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{}
				close(statusCh)
				close(errCh)

				mockClient.On("ContainerWait",
					mock.Anything, "123", mock.Anything,
				).Return((<-chan container.WaitResponse)(statusCh), (<-chan error)(errCh))

				mockClient.On("CopyFromContainer",
					mock.Anything, "123", "/app/test-results",
				).Return(nil, container.PathStat{}, errors.New("no files")).Maybe()

				mockClient.On("ContainerRemove",
					mock.Anything, "123", mock.Anything,
				).Return(nil).Maybe()
			}

			d := &docker{
				logger:           logger,
				config:           cfg,
				client:           mockClient,
				testContainerMap: make(map[string]*entity.ContainerInfo),
				tracer:           tracer,
			}

			id, filesChan, err := d.RunTest(tc.ctx, tc.filename, tc.testID, "userX", "chatY")

			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, id)
				assert.Nil(t, filesChan)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "123", id)
				assert.NotNil(t, filesChan)

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

// createTarArchive creates a tar archive with test files
func createTarArchive(files map[string][]byte) io.ReadCloser {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if _, err := tw.Write(content); err != nil {
			panic(err)
		}
	}
	_ = tw.Close()

	return io.NopCloser(buf)
}

//nolint:funlen
func TestAttachCopyFromContainer(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)
	cfg, _ := config.LoadConfig()
	tracer := otel.Tracer("test")

	tests := []struct {
		name               string
		containerID        string
		waitResp           func() (<-chan container.WaitResponse, <-chan error)
		copyResp           []any
		removeResp         error
		expectedFiles      int
		expectedExtensions []string
	}{
		{
			name:        "success - png and webm files",
			containerID: "container123",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)
				close(errCh)
				return statusCh, errCh
			},
			copyResp: []any{
				createTarArchive(map[string][]byte{
					"screenshot.png": []byte("fake png data"),
					"video.webm":     []byte("fake webm data"),
				}),
				container.PathStat{},
				nil,
			},
			removeResp:         nil,
			expectedFiles:      2,
			expectedExtensions: []string{"png", "webm"},
		},
		{
			name:        "success - only png file",
			containerID: "container456",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)
				close(errCh)
				return statusCh, errCh
			},
			copyResp: []any{
				createTarArchive(map[string][]byte{
					"screenshot.png": []byte("fake png data"),
				}),
				container.PathStat{},
				nil,
			},
			removeResp:         nil,
			expectedFiles:      1,
			expectedExtensions: []string{"png"},
		},
		{
			name:        "success - filters out non-media files",
			containerID: "container789",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)
				close(errCh)
				return statusCh, errCh
			},
			copyResp: []any{
				createTarArchive(map[string][]byte{
					"screenshot.png": []byte("fake png data"),
					"log.txt":        []byte("some log"),
					"test.json":      []byte("{}"),
				}),
				container.PathStat{},
				nil,
			},
			removeResp:         nil,
			expectedFiles:      1,
			expectedExtensions: []string{"png"},
		},
		{
			name:        "error - wait container error",
			containerID: "containerErr1",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				errCh <- errors.New("wait error")
				close(errCh)
				// Don't close statusCh - it will block if no error is read first
				return statusCh, errCh
			},
			removeResp:    nil,
			expectedFiles: 0,
		},
		{
			name:        "error - copy from container fails",
			containerID: "containerErr2",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)
				close(errCh)
				return statusCh, errCh
			},
			copyResp: []any{
				nil,
				container.PathStat{},
				errors.New("copy failed"),
			},
			removeResp:    nil,
			expectedFiles: 0,
		},
		{
			name:        "success - empty archive",
			containerID: "containerEmpty",
			waitResp: func() (<-chan container.WaitResponse, <-chan error) {
				statusCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)
				statusCh <- container.WaitResponse{StatusCode: 0}
				close(statusCh)
				close(errCh)
				return statusCh, errCh
			},
			copyResp: []any{
				createTarArchive(map[string][]byte{}),
				container.PathStat{},
				nil,
			},
			removeResp:    nil,
			expectedFiles: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := mocks.NewMockDockerClient(t)

			statusCh, errCh := tc.waitResp()
			mockClient.On("ContainerWait",
				mock.Anything, tc.containerID, mock.Anything,
			).Return(statusCh, errCh)

			if tc.copyResp != nil {
				mockClient.On("CopyFromContainer",
					mock.Anything, tc.containerID, "/app/test-results",
				).Return(tc.copyResp...)
			}

			mockClient.On("ContainerRemove",
				mock.Anything, tc.containerID, mock.Anything,
			).Return(tc.removeResp)

			d := &docker{
				logger:           logger,
				config:           cfg,
				client:           mockClient,
				testContainerMap: make(map[string]*entity.ContainerInfo),
				tracer:           tracer,
			}

			filesChan := d.attachCopyFromContainer(tc.containerID)

			files := <-filesChan

			assert.Equal(t, tc.expectedFiles, len(files))

			if tc.expectedExtensions != nil {
				extensions := make([]string, len(files))
				for i, file := range files {
					extensions[i] = file.GetFileExtension()
				}
				assert.ElementsMatch(t, tc.expectedExtensions, extensions)
			}

			mockClient.AssertExpectations(t)
		})
	}
}
