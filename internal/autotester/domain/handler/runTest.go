package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/build"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleRunContainer handles the running of the container and returns content of logfile from test
func (a *AutotesterController) HandleRunContainer(c *gin.Context) {
	var params entity.RunTestRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		a.logger.Debug(fmt.Sprintf("currently no request body: err => %s", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if params.UserID == "" || params.TestId == "" || params.SessionID == "" {
		a.logger.Debug("Missing required parameters")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	// find file to mount
	testfile, err := a.saveLocalService.GetTestPath(params.TestId, params.UserID, params.SessionID)
	a.logger.Debug(fmt.Sprintf("Testpath: %s\n", testfile))
	if err != nil {
		a.logger.Debug(fmt.Sprintf("file not available: %s\n", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	if err := runTest(a.logger, c, testfile, a.config.LogDirAutopw); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	// logdir = /var/logs/smart/
	// file is /tmp/userID/sessionID/testcaseID.spec.ts
	LogFilePath := a.config.LogDirAutopw + filepath.Base(testfile) + ".log"
	a.logger.Debug(fmt.Sprintf("Reading log file => %s", LogFilePath))

	// #nosec G304 -- filepath is correct
	content, err := os.ReadFile(LogFilePath)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}
	a.logger.Debug(string(content))
	c.JSON(http.StatusOK, entity.RunTestResponse{
		Result: string(content),
	})
}

func runTest(logger *slog.Logger, ctx context.Context, filename string, logdir string) error {
	if err := assert.NotNil(logger); err != nil {
		return err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Error(fmt.Sprintf("couldnt close client => %s", err.Error()))
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
				Source: logdir,
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

	logger.Debug("Container started, waiting for completion...")

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

	logger.Debug("Container execution completed successfully")
	return nil
}
