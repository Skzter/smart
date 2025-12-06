package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleGetRemoteTestcase retrieves stored test case metadata with optional filters.
// It accepts query parameters for filtering by author, testcaseId, createdAfter, and createdBefore.
// Returns 200 with filtered metadata array, 400 for invalid parameters, or 500 for service errors.
func (a *AutotesterController) HandleGetRemoteTestcase(c *gin.Context) {
	var query entity.GetRemoteTestcaseRequest

	if err := c.ShouldBindQuery(&query); err != nil {
		a.logger.Debug(fmt.Sprintf("GetRemoteTestcase request invalid: %s\n", err.Error()))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	if query.Limit == nil {
		query.Limit = &a.config.DefaultLimitTests
	}
	if query.Offset == nil {
		query.Offset = &a.config.DefaultOffsetTests
	}

	metadata, err := a.remoteTestcaseStorageService.ReadAllMetadataWithFilter(c, &query)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Getting testcase metadata failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, metadata)
}
