package service

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/repository"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Docker handles the execution of tests and reading of the log files
type Docker interface {
	// RunTest starts a container and returns the containerID
	RunTest(ctx context.Context, filename string, testID string) (string, error)

	// AttachToContainer attaches to an already running container
	AttachToContainer(ctx context.Context, containerID string) (*types.HijackedResponse, error)

	// WaitContainer waits for a container to finish
	WaitContainer(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error)

	// GetContainerID retrieves the containerID for a testID
	GetContainerID(testID string) (string, bool)
}

// SSEvent represents a server-sent event with a type and message.
// It is used to communicate events or messages between different parts of the system.
type SSEvent struct {
	Type    string
	Message string
}

// DockerClient is an Interface to interact with a docker client
type DockerClient interface {
	// nolint:lll
	ContainerCreate(ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)

	ContainerStart(ctx context.Context,
		containerID string,
		options container.StartOptions,
	) error

	ContainerWait(ctx context.Context,
		containerID string,
		condition container.WaitCondition,
	) (<-chan container.WaitResponse, <-chan error)

	ContainerAttach(ctx context.Context,
		containerID string,
		options container.AttachOptions,
	) (types.HijackedResponse, error)

	ContainerStop(context.Context,
		string,
		container.StopOptions,
	) error
}

type docker struct {
	logger           *slog.Logger
	config           *config.Config
	filesystem       repository.LogFileSystem
	client           DockerClient
	testContainerMap map[string]string
}

// NewDocker creates a new docker instance
func NewDocker(logger *slog.Logger, config *config.Config, filesystem repository.LogFileSystem, client DockerClient) (Docker, error) {
	if err := assert.NotNil(logger, config, filesystem); err != nil {
		return nil, err
	}
	return &docker{
		logger:           logger,
		config:           config,
		filesystem:       filesystem,
		client:           client,
		testContainerMap: make(map[string]string),
	}, nil
}

// RunTest creates and starts a container for running tests
func (d *docker) RunTest(ctx context.Context, filename string, testID string) (string, error) {
	basefile := path.Base(filename)

	containerConfig := &container.Config{
		Image: "gitlab.dit.htwk-leipzig.de:5050/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/auto-playwright:latest",
		Env:   []string{fmt.Sprintf("OPENAI_API_KEY=%s", build.OpenAIKey)},
		Cmd: []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("cd /apw && npx playwright test %s", basefile),
		},
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  false,
		OpenStdin:    false,
		Tty:          false,
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: "host",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filename,
				Target: fmt.Sprintf("/apw/%s", basefile),
			},
		},
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	d.testContainerMap[testID] = resp.ID
	d.logger.Debug("Container created",
		slog.String("containerID", resp.ID),
		slog.String("testID", testID),
	)

	// Container starten
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Debug("Container started",
		slog.String("containerID", resp.ID),
		slog.String("testID", testID),
	)

	return resp.ID, nil
}

// AttachToContainer attaches to an already running container to stream logs
func (d *docker) AttachToContainer(ctx context.Context, containerID string) (*types.HijackedResponse, error) {
	attachResp, err := d.client.ContainerAttach(ctx, containerID, container.AttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Logs:   true,
	})
	if err != nil {
		return &types.HijackedResponse{}, fmt.Errorf("failed to attach to container: %w", err)
	}

	d.logger.Debug("Attached to container",
		slog.String("containerID", containerID),
	)

	return &attachResp, nil
}

func (d *docker) WaitContainer(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error) {
	statusCh, errCh := d.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	return statusCh, errCh
}

func (d *docker) GetContainerID(testID string) (string, bool) {
	id, found := d.testContainerMap[testID]
	return id, found
}
