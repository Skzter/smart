package service

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"path/filepath"
	"slices"
	"strings"

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
	RunTest(ctx context.Context, filename string, testID, userID, chatID string) (string, <-chan []entity.File, error)
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

	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

type docker struct {
	logger           *slog.Logger
	config           *config.Autotester
	client           DockerClient
	testContainerMap map[string]*entity.ContainerInfo
	tracer           trace.Tracer
}

// NewDocker creates a new docker instance
func NewDocker(logger *slog.Logger, config *config.Autotester, client DockerClient, tracer trace.Tracer) (Docker, error) {
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
func (d *docker) RunTest(ctx context.Context, filename string, testID, userID, chatID string) (string, <-chan []entity.File, error) {
	basefile := path.Base(filename)

	ctx, span := d.tracer.Start(ctx, "docker.RunTest")
	defer span.End()

	containerConfig := &container.Config{
		Image: "gitlab.dit.htwk-leipzig.de:5050/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/auto-playwright:latest",
		Env:   []string{fmt.Sprintf("OPENAI_API_KEY=%s", build.OpenAIKey)},
		Cmd: []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("cd /app && npx playwright test %s", basefile),
		},
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{
		AutoRemove:  false,
		NetworkMode: "host",
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: filename,
				Target: fmt.Sprintf("/app/%s", basefile),
			},
		},
	}

	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create container: %w", err)
	}

	if resp.ID == "" {
		return "", nil, fmt.Errorf("container created without ID")
	}

	d.logger.Debug("Container created",
		slog.String("containerID", resp.ID),
		slog.String("testID", testID),
	)

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", nil, fmt.Errorf("failed to start container: %w", err)
	}

	d.testContainerMap[testID] = &entity.ContainerInfo{
		ContainerID: resp.ID,
		UserID:      userID,
		ChatID:      chatID,
	}

	d.logger.Debug("Container started",
		slog.String("containerID", resp.ID),
		slog.String("testID", testID),
	)

	copyChan := d.attachCopyFromContainer(ctx, resp.ID)

	return resp.ID, copyChan, nil
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

func (d *docker) attachCopyFromContainer(ctx context.Context, containerID string) <-chan []entity.File {
	outchan := make(chan []entity.File)

	statusChan, errChan := d.WaitContainer(context.Background(), containerID)
	go func() {
		defer d.client.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
		files := []entity.File{}
		defer func() { outchan <- files }()

		select {
		case err := <-errChan:
			if err != nil {
				d.logger.Error("derrChan", "err", err)
				return
			}
		case <-statusChan:
		}
		reader, info, err := d.client.CopyFromContainer(context.Background(), containerID, "/app/test-results")
		if err != nil {
			d.logger.Error("error copying from container", "err", err)
			return
		}

		formats := []string{"png", "webm"}

		d.logger.Debug("copied from container", "info", info)
		tr := tar.NewReader(reader)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break // End of archive
			}
			if err != nil {
				d.logger.Error(err.Error())
				continue
			}

			d.logger.Debug("reading file from container", "file", hdr.Name)
			extension, _ := strings.CutPrefix(filepath.Ext(hdr.Name), ".")
			if !slices.Contains(formats, extension) {
				d.logger.Debug(fmt.Sprintf("file %s (extension %s) is not in list %v", hdr.Name, extension, formats))
				continue
			}

			bytes, err := io.ReadAll(tr)
			if err != nil {
				d.logger.Error("error reading file", "file", hdr.Name, "err", err)
				continue
			}
			files = append(files, entity.NewFile(filepath.Base(hdr.Name), bytes, extension))
		}
	}()

	return outchan
}
