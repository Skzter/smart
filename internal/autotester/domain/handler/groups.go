package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
)

// HandleGetGroups returns all groups.
// GET /groups - List all groups
func (a *AutotesterController) HandleGetGroups(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGetGroups")
	defer span.End()

	groups, err := a.groupManager.List(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load group")
		a.metricsService.IncRequestError("load_group_err")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: errors.ErrInternalServer.Error()})
		a.logger.Error("error loading groups", "err", err)
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, groups)
}

// HandleCreateGroup creates a new group.
// POST /groups - Create a new group
func (a *AutotesterController) HandleCreateGroup(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleCreateGroups")
	defer span.End()

	var request entity.CreateGroupRequest

	if err := c.BindJSON(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("error binding request", "err", err)
		return
	}

	groupId, err := a.groupManager.Create(ctx, request.GroupName, request.Description, request.UserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create group")
		a.metricsService.IncRequestError("create_group_err")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: errors.ErrInternalServer.Error()})
		a.logger.Error("error creating group", "err", err)
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.CreateGroupResponse{GroupId: groupId})
}

// HandleAssignChatToGroups assigns a chat to groups.
// POST /chats/:chatId/groups - Assign chat to groups
func (a *AutotesterController) HandleAssignChatToGroups(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleAssignChatToGroups")
	defer span.End()

	var uri entity.AssignChatToGroupRequestURI
	var body entity.AssignChatToGroupRequestJSON

	if err := c.BindUri(&uri); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind URI")
		a.metricsService.IncRequestError("invalid_uri")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("error binding uri", "err", err)
		return
	}

	if err := c.BindJSON(&body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("error binding json", "err", err)
		return
	}

	if err := a.groupManager.AddChatToGroup(ctx, body.GroupId, uri.ChatId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to assign chat to group")
		a.metricsService.IncRequestError("assign_chat_to_group_err")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.Status(http.StatusOK)
}

// HandleRemoveChatFromGroup removes a chat from a group.
// DELETE /chats/:chatId/groups/:groupId - Remove chat from group
func (a *AutotesterController) HandleRemoveChatFromGroup(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleAssignChatToGroups")
	defer span.End()

	var uri entity.RemoveChatFromGroupRequest

	if err := c.BindUri(&uri); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind URI")
		a.metricsService.IncRequestError("invalid_uri")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("error binding uri", "err", err)
		return
	}

	if err := a.groupManager.RemoveChatFromGroup(ctx, uri.GroupId, uri.ChatId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to remove chat from group")
		a.metricsService.IncRequestError("remove_chat_from_group_err")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.Status(http.StatusOK)
}
