package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleRunContainer handles the running of the container and returns content of logfile from test
//
//nolint:funlen
func (a *AutotesterController) HandleRunContainer(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var params entity.RunTestRequest

	_, span := a.tracer.Start(ctx, "autotesterController.HandleRunContainer")
	defer span.End()

	if err := c.ShouldBindJSON(&params); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Debug(fmt.Sprintf("currently no request body: err => %s", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if params.UserID == "" || params.TestId == "" || params.SessionID == "" {
		span.RecordError(errors.New("missing required parameter"))
		span.SetStatus(codes.Error, "missing required parameter")
		a.metricsService.IncRequestError("missing_parameters")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Debug("Missing required parameters")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	// find file to mount
	testfile, err := a.localTestcaseStorageService.GetTestPath(params.TestId, params.UserID, params.SessionID)
	a.logger.Debug(fmt.Sprintf("Testpath: %s\n", testfile))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get test file")
		a.metricsService.IncRequestError("invalid_test_file")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Debug(fmt.Sprintf("file not available: %s\n", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	if err := a.dockerService.RunTest(c, testfile); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to run container")
		a.metricsService.IncRequestError("run_container_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	output, err := a.dockerService.ReadLog(testfile)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "unable to read log of test execution")
		a.metricsService.IncRequestError("read_log_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	// Checks for a line like: "✓  <number> <testId>.spec.ts:<line>:<column>" in the log output.
	pattern := fmt.Sprintf(`(?m)^\s*✓\s+\d+\s+%s:\d+:\d+`, regexp.QuoteMeta(filepath.Base(testfile)))
	passPattern := regexp.MustCompile(pattern)
	if passPattern.MatchString(output) {
		code, err := a.localTestcaseStorageService.Read(params.TestId, params.UserID, params.SessionID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unable to read test file")
			a.metricsService.IncRequestError("read_test_failed")
			a.metricsService.RecordRequestDuration(time.Since(start))
			a.logger.Error(fmt.Sprintf("Reading local test code failed, skipping remote save: %s", err.Error()))
			c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
			return
		}

		test := &entity.TestCase{
			TestID: params.TestId,
			TestCode: entity.TestCode{
				Code: code,
			},
			Status: entity.TestStatusPassed,
		}
		if _, err := a.remoteTestcaseStorageService.SaveTestcase(c, test, params.UserID); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "saving testcase remotely failed")
			a.metricsService.IncRequestError("remote_saving_failed")
			a.logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
			return
		}
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.RunTestResponse{
		Result: output,
	})
}
