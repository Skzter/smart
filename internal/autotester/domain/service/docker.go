package service

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"

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
	RunTest(ctx context.Context, filename string) error
	ReadLog(filename string) (string, error)
}

// DockerClient is an Interface to interact with a docker client
type DockerClient interface {
	// nolint:lll
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error)
}

type docker struct {
	logger     *slog.Logger
	config     *config.Config
	filesystem repository.LogFileSystem
	client     DockerClient
}

// NewDocker creates a new docker instance
func NewDocker(logger *slog.Logger, config *config.Config, filesystem repository.LogFileSystem, client DockerClient) (Docker, error) {
	if err := assert.NotNil(logger, config, filesystem); err != nil {
		return nil, err
	}

	return &docker{
		logger:     logger,
		config:     config,
		filesystem: filesystem,
		client:     client,
	}, nil
}

// RunTest takes the context and filename of the test of the current request and executes the test
func (d *docker) RunTest(ctx context.Context, filename string) error {
	if err := assert.NotNil(ctx); err != nil {
		return err
	}
	if err := assert.StringNotEmpty(filename); err != nil {
		return err
	}

	basefile := path.Base(filename)
	logFileName := basefile + ".log"

	containerConfig := &container.Config{
		Image: "auto-pw",
		Env:   []string{fmt.Sprintf("OPENAI_API_KEY=%s", build.OpenAIKey)},
		Cmd: []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("cd /app && npx playwright test %s --reporter=list | sed 's/\\x1B\\[[0-9;]*[mGKHF]//g' > /logs/%s 2>&1", basefile, logFileName),
		},
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  true,
		NetworkMode: "host",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filename,
				Target: fmt.Sprintf("/app/%s", basefile),
			},
			{
				Type:   mount.TypeBind,
				Source: d.config.LogDirAutopw,
				Target: "/logs",
			},
		},
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Debug("Container started, waiting for completion...")

	statusCh, errCh := d.client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("error waiting for container: %w", err)
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with Status Code: %d and Error: %s", status.StatusCode, status.Error.Message)
		}
	}

	d.logger.Debug("Container execution completed successfully")
	return nil
}

// ReadLog returns the content of the log file from a given test
func (d *docker) ReadLog(filename string) (string, error) {
	// logdir = /var/logs/smart/
	// file is /tmp/userID/sessionID/testcaseID.spec.ts
	LogFilePath := filepath.Base(filename) + ".log"
	d.logger.Debug(fmt.Sprintf("Reading log file => %s", LogFilePath))

	content, err := d.filesystem.ReadFile(LogFilePath)
	if err != nil {
		return "", err
	}
	d.logger.Debug(string(content))
	return string(content), nil
}
