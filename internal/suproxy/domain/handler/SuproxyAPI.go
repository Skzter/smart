package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.POST("/api/v1/Offerlist", getOfferlist)

	err := router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func getOfferlist(c *gin.Context) {
	c.String(http.StatusOK, "This is my Offerlist")
}
