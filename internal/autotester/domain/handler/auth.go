package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleGenerateToken takes a UserID and checks the database for a valid token.
// If there isn't a valid token, it will generate one and return it.
func (a *AutotesterController) HandleGenerateToken(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGenerateToken")
	defer span.End()

	var user entity.User
	if err := c.BindJSON(&user); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_JSON")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}
	if err := assert.StringNotEmpty(user.UserId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing required parameter")
		a.metricsService.IncRequestError("missing_parameters")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("HandleGenerateToken Params", "error", "UserId is empty")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}

	token, err := a.authService.GenerateToken(ctx, user.UserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "generating token failed")
		a.metricsService.IncRequestError("token_generation_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("GenerateToken()", "error", err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{
			Error: "Internal Server Error",
		})
		return
	}
	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, token)
}

// HandleValidateToken validates an opaque bearer token (DB-backed) and returns whether it is valid.
func (a *AutotesterController) HandleValidateToken(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleValidateToken")
	defer span.End()

	token, err := a.authService.GetBearerToken(c.Request.Header)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "missing/invalid authorization header")
		a.metricsService.IncRequestError("missing_or_invalid_auth_header")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	result, err := a.authService.ValidateToken(ctx, token)
	if err != nil || !result.Valid {
		if err != nil {
			span.RecordError(err)
		}
		a.metricsService.IncRequestError("invalid_token")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if result.Revoked {
		a.metricsService.IncRequestError("token_revoked")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.Status(http.StatusOK)
}
