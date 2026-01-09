package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// HandleGenerateToken takes a UserID and checks the database for a valid token.
// If there isn't a valid token, it will generate one and return it.
func (a *AutotesterController) HandleGenerateToken(c *gin.Context) {
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}
	token, err := a.authService.GenerateToken(c.Request.Context(), user.UserId)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{
			Error: "Internal Server Error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
