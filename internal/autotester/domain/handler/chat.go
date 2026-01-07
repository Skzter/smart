package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
	sharedEntity "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/entity"
	sharedErrors "gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/domain/errors"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/shared/lib/assert"
)

// HandleChatRequest processes a chat request from the frontend.
// Expects a JSON with UserRequestDTO and returns a response from the LLM.
//
//nolint:funlen
func (a *AutotesterController) HandleChatRequest(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var userRequest entity.UserRequest

	_, span := a.tracer.Start(ctx, "autotesterController.HandleChatRequest")
	defer span.End()

	if err := c.BindJSON(&userRequest); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
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

	generatedCode, err := a.generationService.GeneratePrompt(c, chat, &userRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to generate prompt")
		a.metricsService.IncRequestError("generation_error")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: err.Error()})
		a.logger.Error("Test generation failed", "error", err)
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK,
		&entity.ResponseForUser{
			Message: sharedEntity.Message{Body: generatedCode},
			UserId:  chat.UserId,
			ChatId:  chat.Id,
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

	c.JSON(http.StatusOK,
		entity.ResponseForUser{
			Message: sharedEntity.Message{},
			UserId:  chat.UserId,
			ChatId:  chat.Id,
		})
}

// HandleUserInfoRequest processes a request for user information.
// Expects a JSON with UserRequestDTO and returns a ResponseForUser.
func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	start := time.Now()
	ctx := c.Request.Context()
	var body entity.UserRequest
	var resp entity.ResponseForUser

	_, span := a.tracer.Start(ctx, "autotesterController.HandleUserInfoRequest")
	defer span.End()

	if err := c.BindJSON(&body); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to bind JSON")
		a.metricsService.IncRequestError("invalid_json")
		a.metricsService.RecordRequestDuration(time.Since(start))
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
	c.JSON(http.StatusOK, entity.ResponseForUser{ChatId: resp.ChatId})
}

// HandleGetUserChats processes a request for all chats of a given user.
// Expects an Auth0-style userId (e.g. auth0|687270280dca20b77cfdcf74) as URL parameter.
func (a *AutotesterController) HandleGetUserChats(c *gin.Context) {
	start := time.Now()
	_, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGetUserChats")
	defer span.End()

	var parameters entity.UserChatSummaryParameters
	if err := c.ShouldBindUri(&parameters); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}
	// checking if style: auth0|id is there
	if !isValid(parameters.UserID) {
		span.RecordError(fmt.Errorf("invalid user id: %s", parameters.UserID))
		span.SetStatus(codes.Error, "invalid user id")
		a.metricsService.IncRequestError("invalid_user_id")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("invalid id format")
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid id format"})
		return
	}

	var queryParameters entity.UserChatSummaryQueryParameters
	if err := c.ShouldBindQuery(&queryParameters); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}

	chats, err := a.chatStorageService.LoadUserChats(c, parameters.UserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load user chats")
		a.metricsService.IncRequestError("load_user_chats_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "no history found for this user"})
		return
	}

	if queryParameters.Limit > len(chats) || queryParameters.Limit == 0 {
		queryParameters.Limit = len(chats)
	}

	chats = chats[:queryParameters.Limit]

	span.SetStatus(codes.Ok, "")
	a.metricsService.IncRequestSuccess()
	a.metricsService.RecordRequestDuration(time.Since(start))
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
	start := time.Now()
	_, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGetUserChats")
	defer span.End()
	var parameters entity.UserChatsParameter

	if err := c.ShouldBindUri(&parameters); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "invalid parameters",
		})
		return
	}

	if !isValid(parameters.UserID) {
		span.RecordError(fmt.Errorf("invalid user id: %s", parameters.UserID))
		span.SetStatus(codes.Error, "invalid user id")
		a.metricsService.IncRequestError("invalid_user_id")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("invalid userId format", "userId", parameters.UserID)
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid userId format"})
		return
	}

	if _, err := uuid.Parse(parameters.ChatID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.metricsService.IncRequestError("invalid_chat_id")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("invalid chatId format", "chatId", parameters.ChatID, "error", err)
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "invalid chatId format"})
		return
	}

	chat, err := a.chatStorageService.LoadChat(c.Request.Context(), parameters.UserID, parameters.ChatID)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrChatNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to find user chat")
			a.metricsService.IncRequestError("find_user_chat_failed")
			a.metricsService.RecordRequestDuration(time.Since(start))
			a.logger.Info("chat not found", "chatId", parameters.ChatID, "userId", parameters.UserID)
			c.JSON(http.StatusNotFound, entity.ErrorMessage{Error: "chat not found"})
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load user chat")
		a.metricsService.IncRequestError("load_user_chat_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("LoadChat failed", "error", err, "chatId", parameters.ChatID, "userId", parameters.UserID)
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "could not load chat"})
		return
	}

	c.JSON(http.StatusOK, chat)
}
