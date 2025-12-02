package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleSaveLocalRequest processes a request to save a test case locally.
// Expects a JSON with LocalSaveRequest and creates a new test case with status NotRun.
// Returns a LocalSaveResponse with the generated test case ID or an error if saving fails.
func (a *AutotesterController) HandleSaveLocalRequest(c *gin.Context) {
	var localSaveRequest entity.LocalSaveRequest

	if err := c.BindJSON(&localSaveRequest); err != nil {
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

	if err := a.localTestcaseStorageService.Save(testcaseToSave, localSaveRequest.UserId, localSaveRequest.ConversationId); err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Saving failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, entity.LocalSaveResponse{TestcaseId: testcaseToSave.TestID, Action: "saved"})
}

// HandleDeleteLocalRequest processes a request to delete a locally stored test case.
// Expects query parameters with testcaseId, userId and conversationId.
// Returns a LocalDeleteResponse confirming the deletion or an error if the deletion fails.
func (a *AutotesterController) HandleDeleteLocalRequest(c *gin.Context) {
	var deleteLocalRequest entity.LocalDeleteRequest

	if err := c.ShouldBindQuery(&deleteLocalRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if deleteLocalRequest.TestcaseId == "" || deleteLocalRequest.UserId == "" || deleteLocalRequest.ConversationId == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	if err := a.localTestcaseStorageService.Delete(deleteLocalRequest.TestcaseId, deleteLocalRequest.UserId, deleteLocalRequest.ConversationId); err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Deleting failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, entity.LocalDeleteResponse{TestcaseId: deleteLocalRequest.TestcaseId, Action: "deleted"})
}
