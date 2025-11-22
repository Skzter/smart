package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleGetTemplate processes a template request from the frontend.
func (a *AutotesterController) HandleGetTemplate(c *gin.Context) {
	if err := assert.StringNotEmpty(a.config.Template); err != nil {
		c.JSON(http.StatusTeapot, "")
		a.logger.Error(err.Error())
		return
	}

	c.JSON(http.StatusOK, entity.Template{Template: a.config.Template})
}
