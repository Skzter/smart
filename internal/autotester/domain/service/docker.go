package service

import (
	"context"
	"fmt"
	"log/slog"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/config"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// Docker handles the execution of tests and reading of the log files
type Docker interface {
	RunTest(ctx context.Context, filename string) error
	ReadLog(filename string) (string, error)
}

type docker struct {
	logger *slog.Logger
	config *config.Config
	logdir string
}

// NewDocker creates a new docker instance
func NewDocker(logger *slog.Logger, config *config.Config) (Docker, error) {
	if err := assert.NotNil(logger, config); err != nil {
		return nil, err
	}
	return &docker{
		logger: logger,
		config: config,
		logdir: config.LogDirAutopw,
	}, nil
}

// RunTest takes the context and filename of the test of the current request and executes the test
func (d *docker) RunTest(ctx context.Context, filename string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer func() {
		if err := cli.Close(); err != nil {
			d.logger.Error(fmt.Sprintf("couldnt close client => %s", err.Error()))
		}
	}()

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
				Source: d.logdir,
				Target: "/logs",
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	d.logger.Debug("Container started, waiting for completion...")

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
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
	return "", nil
}
