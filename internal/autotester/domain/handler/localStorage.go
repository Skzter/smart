package handler

import (
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleSaveLocalRequest processes a request to save a test case locally.
// Expects a JSON with LocalSaveRequest and creates a new test case with status NotRun.
// Returns a LocalSaveResponse with the generated test case ID or an error if saving fails.
func (a *AutotesterController) HandleSaveLocalRequest(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var localSaveRequest entity.LocalSaveRequest

	_, span := a.tracer.Start(ctx, "autotesterController.HandleSaveLocalRequest")
	defer span.End()

	if err := c.BindJSON(&localSaveRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	testcaseToSave := &entity.TestCase{
		TestID: uuid.New().String(),
		TestCode: entity.TestCode{
			Code: localSaveRequest.Code,
		},
		Status: entity.TestStatusNotRun,
	}

	if err := a.localTestcaseStorageService.Save(testcaseToSave, localSaveRequest.UserId, localSaveRequest.ChatId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "saving testcase locally failed")
		a.metricsService.IncRequestError("local_saving_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Saving failed due to internal server error"})
		return
	}

	span.SetStatus(codes.Ok, "successfully saved testcase locally")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.LocalSaveResponse{TestcaseId: testcaseToSave.TestID, Action: "saved"})
}

// HandleDeleteLocalRequest processes a request to delete a locally stored test case.
// Expects query parameters with testcaseId, userId and chatId.
// Returns a LocalDeleteResponse confirming the deletion or an error if the deletion fails.
func (a *AutotesterController) HandleDeleteLocalRequest(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var deleteLocalRequest entity.LocalDeleteRequest

	_, span := a.tracer.Start(ctx, "autotesterController.HandleDeleteLocalRequest")
	defer span.End()

	if err := c.ShouldBindQuery(&deleteLocalRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind form")
		a.metricsService.IncRequestError("invalid_form")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if deleteLocalRequest.TestcaseId == "" || deleteLocalRequest.UserId == "" || deleteLocalRequest.ChatId == "" {
		span.RecordError(errors.New("missing required parameter"))
		span.SetStatus(codes.Error, "missing required parameter")
		a.metricsService.IncRequestError("missing_parameters")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	if err := a.localTestcaseStorageService.Delete(deleteLocalRequest.TestcaseId, deleteLocalRequest.UserId, deleteLocalRequest.ChatId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "deleting testcase locally failed")
		a.metricsService.IncRequestError("local_deletion_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Deleting failed due to internal server error"})
		return
	}

	span.SetStatus(codes.Ok, "successfully deleted testcase locally")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.LocalDeleteResponse{TestcaseId: deleteLocalRequest.TestcaseId, Action: "deleted"})
}
