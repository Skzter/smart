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
	c.String(http.StatusOK, "This is my Offerlist")
}

func postError(c *gin.Context) {
	c.String(http.StatusNotFound, "Error")
}
