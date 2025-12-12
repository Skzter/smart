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
	"go.opentelemetry.io/otel/trace"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Docker interface
type Docker interface {
	RunTest(ctx context.Context, filename string, testID, userID, sessionID string) (string, error)
	AttachToContainer(ctx context.Context, containerID string) (*types.HijackedResponse, error)
	WaitContainer(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error)
	GetContainerInfo(testID string) (*entity.ContainerInfo, bool)
}

// DockerClient interacts with Docker
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
	client           DockerClient
	testContainerMap map[string]*entity.ContainerInfo
	tracer           trace.Tracer
}

// NewDocker creates a new docker instance
func NewDocker(logger *slog.Logger, config *config.Config, client DockerClient, tracer trace.Tracer) (Docker, error) {
	if err := assert.NotNil(logger, config); err != nil {
		return nil, err
	}
	return &docker{
		logger:           logger,
		config:           config,
		client:           client,
		testContainerMap: make(map[string]*entity.ContainerInfo),
		tracer:           tracer,
	}, nil
}

// RunTest creates and starts a container for running tests
func (d *docker) RunTest(ctx context.Context, filename string, testID, userID, sessionID string) (string, error) {
	basefile := path.Base(filename)

	ctx, span := d.tracer.Start(ctx, "docker.RunTest")
	defer span.End()

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

	if resp.ID == "" {
		return "", fmt.Errorf("container created without ID")
	}

	d.logger.Debug("Container created",
		slog.String("containerID", resp.ID),
		slog.String("testID", testID),
	)

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	d.testContainerMap[testID] = &entity.ContainerInfo{
		ContainerID: resp.ID,
		UserID:      userID,
		SessionID:   sessionID,
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

	d.logger.Debug("Attached to container", slog.String("containerID", containerID))
	return &attachResp, nil
}

func (d *docker) WaitContainer(ctx context.Context, containerID string) (<-chan container.WaitResponse, <-chan error) {
	return d.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
}

func (d *docker) GetContainerInfo(testID string) (*entity.ContainerInfo, bool) {
	info, ok := d.testContainerMap[testID]
	return info, ok
}
