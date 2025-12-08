package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
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

	// existence of user needs to be verified, for now just check for empty string
	if assert.StringNotEmpty(userRequest.UserId) != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("No UserID provided in Request")
		return
	}

	chat, err := a.chatManager.LoadChat(c, userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: sharedErrors.ErrInternalServer.Error()})
		a.logger.Error("Loading Chat failed", "error", err)
		return
	}

	defer func() {
		// only update chat when no errors
		if c.Writer.Status() == http.StatusOK {
			if err := a.chatManager.SaveChat(c, chat); err != nil {
				a.logger.Error("Updating stored chat failed", "err", err.Error())
			}
		}
	}()

	valid, msg, err := a.validationService.ValidatePrompt(c, chat, &userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		a.logger.Error("Validation failed", "error", err)
		return
	}

	if !valid {
		c.JSON(http.StatusOK,
			&entity.ResponseForUser{
				Message: sharedEntity.Message{Body: msg},
				UserId:  chat.UserId,
				ChatId:  chat.Id,
			})
		return
	}

	generatedCode, err := a.generationService.GeneratePrompt(c, chat, &userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		a.logger.Error("Test generation failed", "error", err)
		return
	}

	c.JSON(http.StatusOK,
		&entity.ResponseForUser{
			Message: sharedEntity.Message{Body: generatedCode},
			UserId:  chat.UserId,
			ChatId:  chat.Id,
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

// HandleGetUserChats processes a request for all chats of a given user.
// Expects an Auth0-style userId (e.g. auth0|687270280dca20b77cfdcf74) as URL parameter.
func (a *AutotesterController) HandleGetUserChats(c *gin.Context) {
	userID := c.Param("userId")
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
	if len(tokens) != 2 || tokens[0] != "auth0" || tokens[1] == "" {
		return false
	}
	return true
}

// GetChatById returns a full chat including all messages for a given chatId and userId.
func (a *AutotesterController) GetChatById(c *gin.Context) {
	chatID := c.Param("chatId")
	userID := c.Param("userId")

	if !isValid(userID) {
		a.logger.Error("invalid userId format", "userId", userID)
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid userId format"})
		return
	}

	if _, err := uuid.Parse(chatID); err != nil {
		a.logger.Error("invalid chatId format", "chatId", chatID, "error", err)
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid chatId format"})
		return
	}

	chat, err := a.chatStorageService.LoadChat(c.Request.Context(), userID, chatID)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrChatNotFound) {
			a.logger.Info("chat not found", "chatId", chatID, "userId", userID)
			c.JSON(http.StatusNotFound, entity.ErrorMessage{Error: "chat not found"})
			return
		}

		a.logger.Error("LoadChat failed", "error", err, "chatId", chatID, "userId", userID)
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "could not load chat"})
		return
	}

	c.JSON(http.StatusOK, chat)
}
