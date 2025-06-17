package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Structs for Tests
type Offerlist struct {
	Header      string `json:"header"`
	Prompt      string `json:"prompt"`
	Destination string `json:"destination"`
	Request     string `json:"request"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func main() {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", postOfferlist)
		api.POST("/Error", postError)
	}

	err := router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func postOfferlist(c *gin.Context) {
	var offer Offerlist

	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, Error{
			Code:    "400",
			Message: "Invalid JSON",
		})
		return
	}

	// Beispielhafte Rückgabe des Angebots
	c.JSON(http.StatusOK, offer)
}

func postError(c *gin.Context) {
	c.JSON(http.StatusNotFound, Error{
		Code:    "404",
		Message: "Dies ist ein simuliertes Beispiel für einen Fehler",
	})
}

