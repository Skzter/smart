package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleRunContainer startet den Testcontainer und liefert Logs via SSE (über separaten Endpoint)
func (a *AutotesterController) HandleRunContainer(c *gin.Context) {
	var params entity.RunTestRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		a.logger.Debug(fmt.Sprintf("Bad request body: %s", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if params.UserID == "" || params.TestId == "" || params.SessionID == "" {
		a.logger.Debug("Missing required parameters")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Missing required parameters"})
		return
	}

	// Testdatei ermitteln
	testfile, err := a.localTestcaseStorageService.GetTestPath(params.TestId, params.UserID, params.SessionID)
	if err != nil {
		a.logger.Debug(fmt.Sprintf("Test file not found: %s", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	// Container starten (Logs werden im SSE-Handler gelesen)
	_, err = a.dockerService.RunTest(c.Request.Context(), testfile, params.TestId)
	if err != nil {
		a.logger.Error(fmt.Sprintf("Failed to start container: %s", err.Error()))
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		return
	}

	// Direkt zurückmelden: Test läuft
	c.JSON(http.StatusOK, entity.RunTestResponse{
		Result: "Test started",
	})
}
