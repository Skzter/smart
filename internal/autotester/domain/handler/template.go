package handler

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleGetTemplate processes a template request from the frontend.
func (a *AutotesterController) HandleGetTemplate(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()

	_, span := a.tracer.Start(ctx, "autotesterController.HandleGetTemplate")
	defer span.End()

	if err := assert.StringNotEmpty(a.config.Template); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "template is empty")
		a.metricsService.IncRequestError("empty_template")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusTeapot, "")
		a.logger.Error(err.Error())
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.Template{Template: a.config.Template})
}
