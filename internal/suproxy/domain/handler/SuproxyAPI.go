package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/Offerlist", postOfferlist)
		api.POST("/Error", postError)
	}

	err := router.Run("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func postOfferlist(c *gin.Context) {
	var offer map[string]interface{}

	if err := c.ShouldBindJSON(&offer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "400",
			"message": "Invalid JSON",
		})
		return
	}

	c.JSON(http.StatusOK, offer)
}

func postError(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"code":    "404",
		"message": "Dies ist ein simuliertes Beispiel für einen Fehler",
	})
}
