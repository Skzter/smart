package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleRemoteGetTestcase TODO: add godoc
func (a *AutotesterController) HandleRemoteGetTestcase(c *gin.Context) {
	var query entity.GetRemoteTestcaseRequest

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	metadata, err := a.remoteTestcaseStorageService.ReadAllMetadataWithFilter(c, &query)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "Getting testcase metadata failed due to internal server error"})
		return
	}

	c.JSON(http.StatusOK, metadata)
}
