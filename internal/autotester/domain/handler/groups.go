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
		a.logger.Error("error loading chats", "err", err)
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
	c.JSON(http.StatusOK, gin.H{
		"status": []gin.H{},
	})
}

// HandleGetGroupChats returns all chats of a group.
// GET /groups/:groupId/chats - Get all chats of a group
func (a *AutotesterController) HandleGetGroupChats(c *gin.Context) {
	_ = c.Param("groupId")
	c.JSON(http.StatusOK, gin.H{
		"chats": []gin.H{},
	})
}

// HandleAssignChatToGroups assigns a chat to groups.
// POST /chats/:chatId/groups - Assign chat to groups
func (a *AutotesterController) HandleAssignChatToGroups(c *gin.Context) {
	_ = c.Param("chatId")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chat assigned to groups",
	})
}

// HandleRemoveChatFromGroup removes a chat from a group.
// DELETE /chats/:chatId/groups/:groupId - Remove chat from group
func (a *AutotesterController) HandleRemoveChatFromGroup(c *gin.Context) {
	_ = c.Param("chatId")
	_ = c.Param("groupId")
	c.JSON(http.StatusOK, gin.H{
		"message": "Chat removed from group",
	})
}
