package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
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

	if err := a.dockerService.RunTest(c, testfile); err != nil {
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
