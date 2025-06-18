package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "Hello World")
	})

	if err := r.Run(":8081"); err != nil {
		panic(err)
	}
}
