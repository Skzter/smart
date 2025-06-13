package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/entity"
)

// handle Frontend Request JSON to String
func frontendToBackend(c *gin.Context) {
	var body entity.ChatRequest

	if err := c.BindJSON(&body); err != nil {
		return
	}
	c.IndentedJSON(http.StatusCreated, body)

	// Call Request to openAI
	NewRequest(body.Message, body.ConversationID)
}

func backendResponse(c *gin.Context) {

}

func main() {
	router := gin.Default()
	router.POST("/api/v1/chat", frontendToBackend)
	router.Run("localhost:8080")
}
