package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
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

// HandleChatRequestValidity processes a chat request for prompt validity checking.
// Expects a JSON with UserRequestDTO and returns whether the prompt is valid or not.
func (a *AutotesterController) HandleChatRequestValidity(c *gin.Context) {
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
				Message: sharedEntity.Message{Body: msg}, // show reason for failure
				UserId:  userRequest.UserId,
				ChatId:  userRequest.ChatId,
			})
		return
	}

	c.JSON(http.StatusOK,
		&entity.ResponseForUser{
			Message: sharedEntity.Message{Body: "Prompt validated successfully!"},
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
	userID := c.Param("UserID")
	// checking if style: auth0|id is there
	if !isValid(userID) {
		a.logger.Error("invalid id format")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid id format"})
		return
	}

	limitStr := c.Query("limit")
	if limitStr == "" {
		limitStr = "0"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "limit has to be a number"})
		return
	}

	if limit < 0 {
		a.logger.Error(fmt.Sprintf("%d, limit has to be greater than zero", limit))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: fmt.Sprintf("%d, limit has to be greater than zero", limit)})
		return
	}

	chats, err := a.chatStorageService.LoadUserChats(c, userID)
	if err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "no history found for this user"})
		return
	}

	if limit > len(chats) || limit == 0 {
		limit = len(chats)
	}

	chats = chats[:limit]

	c.JSON(http.StatusOK, entity.ChatSummarys{
		ChatSummarys: chats,
	})
}

func isValid(userId string) bool {
	tokens := strings.Split(userId, "|")
	if len(tokens) < 2 || tokens[0] != "auth0" || tokens[1] == "" {
		return false
	}
	return true
}
