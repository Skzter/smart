package handler

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

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

	// can only run on container a time - needs fix in sprint 5
	// docker pkg should fix this
	// #nosec G204 - used as quick solution, later docker pkg
	cmd := exec.Command("run-container.sh", fmt.Sprintf("TEST_FILE=%s", testfile))
	a.logger.Debug(fmt.Sprintf("cmd line: %s\n", cmd))
	a.logger.Debug("Starting the task")
	if err := cmd.Run(); err != nil {
		a.logger.Debug(fmt.Sprintf("Command execution error: %s\n", err.Error()))
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	// logdir = logs/
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

	resultStr := string(content)
	pattern := fmt.Sprintf(`(?m)^\s*✓\s+\d+\s+%s:\d+:\d+`, regexp.QuoteMeta(filepath.Base(testfile)))
	passPattern := regexp.MustCompile(pattern)
	if passPattern.MatchString(resultStr) {
		code, err := a.saveLocalService.Read(params.TestId, params.UserID, params.SessionID)
		if err != nil {
			a.logger.Error(err.Error())
		}

		test := &entity.TestCase{
			TestID: params.TestId,
			TestCode: entity.TestCode{
				Code: code,
			},
			Status: entity.TestStatusPassed,
		}
		if err := a.saveTestRemoteServcie.SaveTestCase(c, test); err != nil {
			a.logger.Error(err.Error())
		}
	}

	c.JSON(http.StatusOK, entity.RunTestResponse{
		Result: resultStr,
	})
}
