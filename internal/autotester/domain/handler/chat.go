package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleChatRequest processes a chat request from the frontend.
// Expects a JSON with UserRequestDTO and returns a response from the LLM.
func (a *AutotesterController) HandleChatRequest(c *gin.Context) {
	var userRequest entity.UserRequest

	if err := c.BindJSON(&userRequest); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("JSON binding failed", "error", err)
		return
	}

	if userRequest.ChatId == "" {
		userRequest.ChatId = uuid.New().String()
	}

	valid, msg, err := a.validationService.ValidatePrompt(c, userRequest.Message.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		a.logger.Error("Validation failed", "error", err)
		return
	}

	if !valid {
		c.JSON(http.StatusOK,
			&entity.ResponseForUser{
				Message: sharedEntity.Message{Body: msg},
				UserId:  userRequest.UserId,
				ChatId:  userRequest.ChatId,
			})
		return
	}

	generatedCode, err := a.generationService.GeneratePrompt(c, userRequest.Message.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		a.logger.Error("Test generation failed", "error", err)
		return
	}

	c.JSON(http.StatusOK,
		&entity.ResponseForUser{
			Message: sharedEntity.Message{Body: generatedCode},
			UserId:  userRequest.UserId,
			ChatId:  userRequest.ChatId,
		})
}

// HandleUserInfoRequest processes a request for user information.
// Expects a JSON with UserRequestDTO and returns a ResponseForUser.
func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequest
	var resp entity.ResponseForUser
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.JSON(http.StatusOK, entity.ResponseForUser{ChatId: resp.ChatId})
}

// HandleGetUserChats processes a request for all chats of a given user
// Expects a valid uuid as url parameter
// valid example: /chats/0bc024d1-5e82-435b-8b2e-dc88493a8a28
// invalid example: /chats/1234 or /chats/hahahihi
func (a *AutotesterController) HandleGetUserChats(c *gin.Context) {
	paramID := c.Param("UserID")
	// nach einigen tests sehr sicher, dass das hier eig nie failen wird da man keine request an die url schicken kann ohne id
	// failt einfach direkt weil 404 not found
	if err := assert.StringNotEmpty(paramID); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	userUUID, err := uuid.Parse(paramID)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: err.Error()})
		return
	}

	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = a.config.DefaultLimitChats
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "limit has to be a number"})
		return
	}

	if limit <= 0 {
		a.logger.Error(fmt.Sprintf("%d, limit has to be greater than zero", limit))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: fmt.Sprintf("%d, limit has to be greater than zero", limit)})
		return
	}

	c.JSON(http.StatusOK, userUUID)
}
