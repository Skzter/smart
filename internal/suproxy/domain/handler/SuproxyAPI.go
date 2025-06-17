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
	}

	err := router.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}

func postOfferlist(c *gin.Context) {
	c.String(http.StatusOK, "This is my Offerlist")
}
