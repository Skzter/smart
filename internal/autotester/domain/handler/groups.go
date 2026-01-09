package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleGetGroups returns all groups.
// GET /groups - List all groups
func (a *AutotesterController) HandleGetGroups(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"groups": []gin.H{},
	})
}

// HandleCreateGroup creates a new group.
// POST /groups - Create a new group
func (a *AutotesterController) HandleCreateGroup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Group created",
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
