package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"

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
	var userRequest entity.UserRequest

	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleChatRequest")
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

	chat, err := a.chatManager.LoadChat(ctx, userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: sharedErrors.ErrInternalServer.Error()})
		a.logger.Error("Loading Chat failed", "error", err)
		return
	}

	defer func() {
		// only update chat when no errors
		if c.Writer.Status() == http.StatusOK {
			if err := a.chatManager.SaveChat(ctx, chat, userRequest.UserId); err != nil {
				a.logger.Error("Updating stored chat failed", "err", err.Error())
			}
		}
	}()

	generatedCode, err := a.generationService.GeneratePrompt(ctx, chat, &userRequest)
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
			UserId:  userRequest.UserId,
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
			if err := a.chatManager.SaveChat(c, chat, userRequest.UserId); err != nil {
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
				UserId:  userRequest.UserId,
				ChatId:  chat.Id,
			})
		return
	}

	c.JSON(http.StatusOK,
		entity.ResponseForUser{
			Message: sharedEntity.Message{},
			UserId:  userRequest.UserId,
			ChatId:  chat.Id,
		})
}

// HandleUserInfoRequest processes a request for user information.
// Expects a JSON with UserRequestDTO and returns a ResponseForUser.
func (a *AutotesterController) HandleUserInfoRequest(c *gin.Context) {
	start := time.Now()
	var body entity.UserRequest
	var resp entity.ResponseForUser

	_, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleUserInfoRequest")
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

// HandleGetChats processes a request for all chats, optionally filtered by groups.
func (a *AutotesterController) HandleGetChats(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGetUserChats")
	defer span.End()

	var queryParameters entity.UserChatSummaryQueryParameters
	if err := c.BindQuery(&queryParameters); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "Bad Request",
		})
		return
	}

	chats, err := a.chatStorageService.LoadSummaries(ctx, queryParameters.Groups...)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load user chats")
		a.metricsService.IncRequestError("load_user_chats_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "failed to load chats"})
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

// GetChatById returns a full chat including all messages for a given chatId and userId.
func (a *AutotesterController) GetChatById(c *gin.Context) {
	start := time.Now()
	ctx, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleGetUserChats")
	defer span.End()

	var parameters entity.UserChatsParameter

	if err := c.BindUri(&parameters); err != nil {
		a.logger.Error(err.Error())
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{
			Error: "invalid parameters",
		})
		return
	}

	chat, err := a.chatStorageService.LoadChat(ctx, parameters.ChatID)
	if err != nil {
		if errors.Is(err, sharedErrors.ErrChatNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to find user chat")
			a.metricsService.IncRequestError("find_user_chat_failed")
			a.metricsService.RecordRequestDuration(time.Since(start))
			a.logger.Info("chat not found", "chatId", parameters.ChatID)
			c.JSON(http.StatusNotFound, entity.ErrorMessage{Error: "chat not found"})
			return
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to load user chat")
		a.metricsService.IncRequestError("load_user_chat_failed")
		a.metricsService.RecordRequestDuration(time.Since(start))
		a.logger.Error("LoadChat failed", "error", err, "chatId", parameters.ChatID)
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "could not load chat"})
		return
	}

	c.JSON(http.StatusOK, chat)
}

// HandleUpdateChatTitle allows a user to update the title of an existing chat.
func (a *AutotesterController) HandleUpdateChatTitle(c *gin.Context) {
	_, span := a.tracer.Start(c.Request.Context(), "autotesterController.HandleUpdateChatTitle")
	defer span.End()

	chatID := c.Param("chatId")
	userID := c.Param("userId")

	if assert.StringNotEmpty(userID) != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		a.logger.Error("No UserID provided in Request")
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := c.BindJSON(&req); err != nil {
		a.logger.Error("JSON binding failed", "error", err)
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	chat, err := a.chatStorageService.LoadChat(c.Request.Context(), chatID)
	if err != nil {
		a.logger.Error("LoadChat failed", "error", err, "chatId", chatID, "userId", userID)
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "could not load chat"})
		return
	}

	if chat.Author != userID {
		c.JSON(http.StatusForbidden, entity.ErrorMessage{Error: "Unauthorized"})
		return
	}

	title := strings.TrimSpace(req.Title)
	if len(title) == 0 || len(title) > 30 {
		c.JSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Title must be 1–30 characters"})
		return
	}

	chat.Title = title

	if err := a.chatManager.SaveChat(c.Request.Context(), chat, userID); err != nil {
		a.logger.Error("Saving updated chat failed", "error", err)
		c.JSON(http.StatusInternalServerError, entity.ErrorMessage{Error: "could not save chat"})
		return
	}

	a.metricsService.IncRequestSuccess()
	span.SetStatus(codes.Ok, "")

	c.JSON(http.StatusOK, chat)
}
