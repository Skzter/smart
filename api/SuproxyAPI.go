package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/", homepage)

	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func homepage(c *gin.Context) {
	c.String(http.StatusOK, "This is my Homepage")
}
