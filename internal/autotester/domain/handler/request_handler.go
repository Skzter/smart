package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// handle Frontend Request JSON to String
func HandleChatRequest(c *gin.Context) {
	var body entity.ChatRequest
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}

	// validation Method

	// Call Request to openAI
	// service needed
	// entity.Request(body.Message, body.ConversationID)

	// handling response openai

	// respons to frontend
	c.IndentedJSON(http.StatusOK, entity.ChatResponse{Content: body})
}

func HandleUserInfoRequest(c *gin.Context) {
	var body entity.UserRequest
	if err := c.BindJSON(&body); err != nil {
		c.IndentedJSON(http.StatusBadRequest, entity.ErrorMessage{Error: "Bad Request"})
		return
	}
	c.IndentedJSON(http.StatusOK, entity.UserResponse{UserID: body.UserID, Sessions: nil})
}

func main() {
	router := gin.Default()
	router.POST("/api/v1/chat", HandleChatRequest)
	router.POST("/api/v1/userInfo", HandleUserInfoRequest)
	err := router.Run("localhost:8080")

	if err != nil {
		return
	}
}
