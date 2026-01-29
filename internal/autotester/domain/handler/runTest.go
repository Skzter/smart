package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleRunContainer validates the request, resolves the test file, and starts a Docker-based test run. Uploads any debugging media artifacts outside of request context.
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

	if params.UserID == "" || params.TestId == "" || params.ChatID == "" {
		span.RecordError(errors.New("missing required parameter"))
		span.SetStatus(codes.Error, "missing required parameter")
		a.metricsService.IncRequestError("missing_parameters")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Debug("Missing required parameters")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	testfile, err := a.localTestcaseStorageService.GetTestPath(params.TestId, params.UserID, params.ChatID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get test file")
		a.metricsService.IncRequestError("invalid_test_file")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Debug(fmt.Sprintf("file not available: %s\n", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	_, filesChan, err := a.dockerService.RunTest(c.Request.Context(), testfile, params.TestId, params.UserID, params.ChatID)
	if err != nil {
		span.SetStatus(codes.Error, "unable to run container")
		a.metricsService.IncRequestError("run_container_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	go func() {
		files, ok := <-filesChan
		if ok {
			a.logger.Debug("recieved files from container", "count", len(files))
			for _, file := range files {
				if err := a.mediaStorageService.UploadMedia(context.Background(), params.TestId, file); err != nil {
					a.logger.Error("error uploading media",
						"testId", params.TestId,
						"fileName", file.GetFileName(),
						"err", err)
				}
			}
		} else {
			a.logger.Debug("recieved no files")
		}
	}()

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.RunTestResponse{
		Result: "Test started",
	})
}
