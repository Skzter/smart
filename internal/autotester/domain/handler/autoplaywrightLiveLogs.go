package handler

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/service"
)

//nolint:gochecknoglobals // test helper – allows injection in unit tests
var (
	// PipeFactory allows tests to override io.Pipe
	PipeFactory = pipeFactory

	// StdcopyFunc allows tests to override stdcopy.StdCopy
	StdcopyFunc = stdcopyFunc

	// CloseFunc allows tests to override io.Closer.Close
	CloseFunc = closeFunc
)

func pipeFactory() (*io.PipeReader, *io.PipeWriter) {
	return io.Pipe()
}

func stdcopyFunc(dstout, dsterr io.Writer, src io.Reader) (int64, error) {
	return stdcopy.StdCopy(dstout, dsterr, src)
}

func closeFunc(c io.Closer) error { return c.Close() }

func safeSend(ch chan<- service.SSEvent, ev service.SSEvent) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while sending event: %v", r)
		}
	}()

	select {
	case ch <- ev:
		return nil
	default:
		return fmt.Errorf("failed to send SSEvent: channel blocked or closed")
	}
}

// HandleLogRequest handles the log streaming request for a specific test container.
func (a *AutotesterController) HandleLogRequest(c *gin.Context) {
	testID := c.Param("testId")

	containerID, containerIDExists := a.dockerService.GetContainerID(testID)
	if !containerIDExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found - test may not be running"})
		return
	}

	attachResp, err := a.dockerService.AttachToContainer(c, containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach to container"})
		return
	}
	defer attachResp.Close()

	sseChan := make(chan service.SSEvent, 100)

	go func() {
		defer close(sseChan)

		stdoutReader, stdoutWriter := PipeFactory()
		stderrReader, stderrWriter := PipeFactory()

		go func() {
			_, stdcopyErr := StdcopyFunc(stdoutWriter, stderrWriter, attachResp.Reader)

			if err := CloseFunc(stdoutWriter); err != nil {
				_ = safeSend(sseChan, service.SSEvent{
					Type:    "error",
					Message: "Failed to close stdoutWriter: " + err.Error(),
				})
			}
			if err := CloseFunc(stderrWriter); err != nil {
				_ = safeSend(sseChan, service.SSEvent{
					Type:    "error",
					Message: "Failed to close stderrWriter: " + err.Error(),
				})
			}

			if stdcopyErr != nil && stdcopyErr != io.EOF {
				_ = safeSend(sseChan, service.SSEvent{
					Type:    "error",
					Message: stdcopyErr.Error(),
				})
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stdoutReader)
			for scanner.Scan() {
				_ = safeSend(sseChan, service.SSEvent{
					Type:    "log",
					Message: scanner.Text(),
				})
			}
		}()

		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			_ = safeSend(sseChan, service.SSEvent{
				Type:    "log",
				Message: scanner.Text(),
			})
		}
	}()

	// Container status channels
	statusCh, errCh := a.dockerService.WaitContainer(c.Request.Context(), containerID)

	c.Stream(func(w io.Writer) bool {
		select {
		case ev, ok := <-sseChan:
			if !ok {
				return true
			}
			c.SSEvent(ev.Type, ev.Message)
			return true

		case status := <-statusCh:
			c.SSEvent("status", status)
			c.SSEvent("finished", "container exited")
			return false

		case err := <-errCh:
			c.SSEvent("error", err)
			c.SSEvent("error", "container errored")
			return false

		case <-c.Request.Context().Done():
			return false
		}
	})
}
